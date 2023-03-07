/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/COSAE-FR/ripugw/collect"
	"github.com/COSAE-FR/ripugw/conf"
	"github.com/COSAE-FR/ripugw/inform"
	"github.com/JulesMike/speedtest"
	log "github.com/sirupsen/logrus"
	"gopkg.in/hlandau/easyconfig.v1"
	"gopkg.in/hlandau/service.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const Version = "1.1.5"

var httpClient = &http.Client{
	Transport: &http.Transport{
		DisableCompression: true,
		DisableKeepAlives:  true,
	},
}

func SendInform(message *inform.Inform, config *conf.Config) (inform.Message, error) {
	key := config.Management.GetKey()
	logger := config.Log.WithFields(log.Fields{
		"component":   "send_inform",
		"use_aes_gcm": fmt.Sprintf("%v", config.Management.UseAesGcm),
		"default_key": fmt.Sprintf("%v", key.IsDefault()),
		"mac":         message.Mac.String(),
	})
	mode := config.Management.GetCryptoMode()
	logger.Debug("Preparing packet")
	p := inform.NewPacket(message.Mac, message, key, mode)

	body, err := p.Marshal()
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(body)
	req, err := http.NewRequest("POST", message.InformUrl, r)
	if err != nil {
		logger.Errorf("Cannot prepare POST request: %v", err)
		return nil, err
	}

	addr, err := url.Parse(message.InformUrl)
	if err != nil {
		logger.Errorf("Cannot parse Inform URL: %v", err)
		return nil, err
	}

	req.Host = addr.Hostname()
	setHeader(req, "user-agent", "AirControl Agent v1.0")
	setHeader(req, "content-type", "application/x-binary")
	setHeader(req, "content-length", strconv.Itoa(len(body)))

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("Cannot send POST request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "application/x-binary" {
		logger.Errorf("Received status code: %d with CT: %s", resp.StatusCode, resp.Header.Get("Content-Type"))
		return inform.ResponseFromHttpCode(resp.StatusCode), nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Cannot read server response: %v", err)
		return nil, err
	}

	rPacket := &inform.Packet{}

	err = rPacket.Unmarshal(data, func(ap inform.HardwareAddr) (inform.Key, error) {
		return key, nil
	})
	if err != nil {
		logger.Errorf("Error unmarshalling response: %+v", err)
		return nil, err
	}

	informResp, ok := rPacket.Msg.(inform.InformResponse)
	if !ok {
		logger.Errorf("Invalid Inform response")
		return nil, errors.New("invalid")
	}
	return informResp, nil
}

func SpeedTest(svc *Service) {
	logger := svc.Log.WithField("component", "speed_test")
	logger.Debug("Launching speed_test handler")
	client, err := speedtest.NewDefaultClient()
	if err != nil {
		logger.Errorf("error creating client: %v", err)
		return
	}
	server, err := client.GetServer("")
	if err != nil {
		logger.Errorf("error getting server: %v", err)
		return
	}
	dstatus := 1
	dmbps, err := client.Download(server)
	if err != nil {
		logger.Errorf("error getting download speed: %v", err)
		dstatus = 0
	}

	ustatus := 1
	umbps, err := client.Upload(server)
	if err != nil {
		logger.Errorf("error getting upload speed: %v", err)
		ustatus = 0
	}
	pstatus := 1
	if server.Latency == 0 {
		pstatus = 0
	}
	svc.Lock.Lock()
	svc.SpeedTest = &inform.SpeedTestStatus{
		Latency:        uint64(server.Latency),
		RunDate:        uint64(time.Now().Unix()),
		RunTime:        uint64(time.Now().Unix()),
		StatusDownload: uint64(dstatus),
		StatusPing:     uint64(pstatus),
		StatusUpload:   uint64(ustatus),
		XputDownload:   dmbps,
		XputUpload:     umbps,
	}
	svc.Lock.Unlock()
	if err := svc.Write(); err != nil {
		logger.Errorf("cannot write configuration: %v", err)
	} else {
		logger.Debugf("Last speed test written ton configuration file")
	}

}

func setHeader(r *http.Request, key, value string) {
	r.Header.Del(key)
	r.Header.Set(key, value)
}

const AppName = "ripugw"

type Service struct {
	*conf.Config
	InformTicker *time.Ticker
	InformStop   chan bool
}

func New(cfg ServiceConfig) (*Service, error) {
	var err error
	configuration, err := conf.New(cfg.File, cfg.Json)
	svc := Service{Config: configuration}
	svc.Log.WithFields(log.Fields{
		"component": "daemon_creator",
		"version":   Version,
	}).Debug("Starting daemon")
	return &svc, err
}

func informTick(svc *Service) {
	logger := svc.Log.WithField("component", "inform")
	logger.Info("Launching Inform handler")
	for {
		select {
		case <-svc.InformStop:
			logger.Info("Stopping Inform handler")
			return
		case <-svc.InformTicker.C:
			logger.Debug("Inform tick")
			configVersion := "0123456789abcdef"
			if len(svc.Management.Version) > 0 {
				configVersion = svc.Management.Version
			}

			var informPacket inform.Inform
			var err error
			if svc.PfSenseMode {
				informPacket, err = collect.RequestFromPfsense(svc.General.Url, configVersion, *svc.PfSense, *svc.PfSenseInterfaces, svc.SpeedTest)
				if err != nil {
					logger.Errorf("Cannot prepare pfSense inform packet: %s", err)
					continue
				}
			} else {
				informPacket, err = collect.Request(svc.General.Url, configVersion)
				if err != nil {
					logger.Errorf("Cannot prepare inform packet: %s", err)
					continue
				}
			}
			if svc.General.LogLevel == "trace" {
				packet, err := json.MarshalIndent(informPacket, "", "\t")
				if err != nil {
					logger.Tracef("Cannot marshal Inform packet: %+v", err)
				} else {
					logger.Tracef("Packet to send: \n %s", packet)
				}
			}
			resp, err := SendInform(&informPacket, svc.Config)
			if err != nil {
				logger.Errorf("Cannot send inform packet: %s", err)
				continue
			}
			r, _ := json.Marshal(resp)
			logger.Tracef("Received: %s", r)
			switch response := resp.(type) {
			case *inform.SetParam:
				keyChange, versionChange, cryptoModeChange := false, false, false

				key, keyChange := response.ManagementConfig["authkey"]
				if keyChange {
					logger.Debug("Key changed!")
					svc.Config.Management.Key = key
				}
				version, versionChange := response.ManagementConfig["cfgversion"]
				if versionChange {
					logger.Debug("Version changed!")
					svc.Config.Management.Version = version
				}
				cryptoMode, cryptoModeChange := response.ManagementConfig["use_aes_gcm"]
				if cryptoModeChange {
					logger.Debug("AES mode changed!")
					svc.Config.Management.UseAesGcm = cryptoMode == "true"
				}
				if keyChange || versionChange || cryptoModeChange {
					logger.Debugf("Decoded response authkey: %s, default: %v", svc.Config.Management.Key, svc.Config.Management.GetKey().IsDefault())
					_ = svc.Config.Write()
				}
			case *inform.Noop:
				logger.Debugf("Received Noop message")
			case *inform.Cmd:
				logger.Debugf("Received Cmd message")
				switch response.Command {
				case "speed-test":
					logger.Debugf("Command type: %s", response.Command)
					go SpeedTest(svc)
				default:
					logger.Debugf("Unknown command: %s", response.Command)

				}
			}
		}
	}
}

func (s *Service) Start() error {
	logger := s.Log.WithField("component", "start_handler")

	logger.Debugf("Configuring Inform interval: %d", s.General.InformInterval)
	informDuration := time.Duration(s.General.InformInterval) * time.Second

	logger.Debug("Creating Inform handler")
	s.InformTicker = time.NewTicker(informDuration)
	s.InformStop = make(chan bool)

	if s.SpeedTest == nil || s.SpeedTest.RunTime == 0 {
		logger.Debug("Starting speed_test")
		go SpeedTest(s)
	}

	logger.Debug("Starting Inform handler")
	go informTick(s)
	return nil
}

func (s Service) Stop() error {
	logger := s.Log.WithField("component", "stop_handler")
	logger.Info("Stopping service")
	logger.Debug("Stopping inform")
	s.InformTicker.Stop()
	s.InformStop <- true
	return nil
}

func main() {
	// Early log settings
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.ErrorLevel)

	logger := log.WithFields(log.Fields{
		"app":       AppName,
		"component": "main",
	})
	cfg := ServiceConfig{}

	configurator := &easyconfig.Configurator{
		ProgramName: "ugw",
	}

	err := easyconfig.Parse(configurator, &cfg)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	if cfg.File == "" {
		cfg.File = defaultConfigFile
	}
	logger.Debugf("Started with %#v", cfg)
	service.Main(&service.Info{
		Name:      AppName,
		AllowRoot: true,
		NewFunc: func() (service.Runnable, error) {
			return New(cfg)
		},
	})
}
