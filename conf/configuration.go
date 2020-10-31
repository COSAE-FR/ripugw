/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package conf

import (
	"encoding/json"
	"encoding/xml"
	"github.com/BurntSushi/toml"
	"github.com/COSAE-FR/ripugw/collect"
	"github.com/COSAE-FR/ripugw/inform"
	"github.com/COSAE-FR/ripugw/pfconf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
)

const (
	defaultInformUrl      = "http://unifi:8080/inform"
	defaultInformInterval = 15
)

type Config struct {
	General           general                          `toml:"general" json:"general"`
	Management        Management                       `toml:"mgmt_cfg" json:"mgmt_cfg"`
	PfSenseInterfaces *collect.PfSenseTranslationTable `toml:"pfsense_interfaces" json:"pfsense_interfaces"`
	path              string                           `toml:"-" json:"-"`
	useJson           bool                             `toml:"-" json:"-"`
	Log               *log.Entry                       `toml:"-" json:"-"`
	PfSenseMode       bool                             `toml:"-" json:"-"`
	PfSense           *pfconf.Configuration            `toml:"-" json:"-"`
	SpeedTest         *inform.SpeedTestStatus          `toml:"last_speedtest,omitempty" json:"last_speedtest,omitempty"`
	Lock              sync.Mutex                       `toml:"-" json:"-"`
}

type general struct {
	Url            string   `toml:"url" json:"url"`
	Adopted        bool     `toml:"adopted" json:"adopted"`
	LogLevel       string   `toml:"log_level" json:"log_level"`
	LogFile        string   `toml:"log_file" json:"log_file"`
	PfSenseXml     string   `toml:"pfsense_xml" json:"pfsense_xml"`
	InformInterval int      `toml:"interval" json:"interval"`
	LogFileWriter  *os.File `toml:"-" json:"-"`
}

type Management struct {
	Version   string `toml:"configversion" json:"configversion"`
	UseAesGcm bool   `toml:"use_aes_gcm" json:"use_aes_gcm"`
	Key       string `toml:"authkey" json:"authkey"`
}

func (c *Config) Read() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if c.useJson {
		return c.readJson()
	} else {
		return c.readToml()
	}
}

func (c *Config) readToml() error {
	if _, err := toml.DecodeFile(c.path, c); err != nil {
		return err
	}
	return nil
}

func (c *Config) readJson() error {
	jsonFile, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(byteValue, c)
}

func (c *Config) Write() error {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.useJson {
		return c.writeJson()
	} else {
		return c.writeToml()
	}
}

func (c *Config) writeToml() error {
	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func (c *Config) writeJson() error {
	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(c); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func (c *Config) setUpLog() {
	if len(c.General.LogLevel) == 0 {
		c.General.LogLevel = "error"
	}
	logLevel, err := log.ParseLevel(c.General.LogLevel)
	if err != nil {
		logLevel = log.ErrorLevel
	}
	log.SetLevel(logLevel)
	logger := log.WithFields(log.Fields{
		"app":       "ripugw",
		"component": "config_loader",
	})
	if len(c.General.LogFile) > 0 {
		f, err := os.OpenFile(c.General.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err == nil {
			c.General.LogFileWriter = f
			log.SetOutput(f)
		} else {
			logger.Errorf("Cannot open log file %s. Logging to stderr.", c.General.LogFile)
		}
	} else {
		c.General.LogFileWriter = os.Stderr
	}
	c.Log = logger
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (c *Config) check() (changed bool, err error) {
	logger := c.Log.WithField("component", "config_checker")
	if len(c.General.Url) == 0 {
		logger.Warnf("no inform URL in configuration file, default used: %s", defaultInformUrl)
		c.General.Url = defaultInformUrl
		changed = true
	}
	if c.General.InformInterval == 0 {
		logger.Warnf("no inform interval in configuration file, default used: %d", defaultInformInterval)
		c.General.InformInterval = defaultInformInterval
		changed = true
	}
	if len(c.General.PfSenseXml) > 0 && fileExists(c.General.PfSenseXml) {
		if len(c.PfSenseInterfaces.Wan) == 0 || len(c.PfSenseInterfaces.Lan) == 0 {
			logger.Warn("no interface translation table between pfSense and physical interfaces")
		} else {
			xmlFile, err := os.Open(c.General.PfSenseXml)
			// if we os.Open returns an error then handle it
			if err != nil {
				logger.Errorf("cannot open pfSense configuration file %s: %s", c.General.PfSenseXml, err)
			} else {
				logger.Info("Successfully opened pfSense configuration file")
				// defer the closing of our xmlFile so that we can parse it later on
				defer func() {
					err = xmlFile.Close()
					if err != nil {
						logger.Errorf("cannot close pfSense configuration file: %s", c.General.PfSenseXml)
					}
				}()
				byteValue, err := ioutil.ReadAll(xmlFile)
				if err != nil {
					logger.Errorf("cannot read pfSense configuration file %s: %s", c.General.PfSenseXml, err)
				} else {
					if err := xml.Unmarshal(byteValue, &c.PfSense); err != nil {
						logger.Errorf("cannot parse pfSense configuration file %s: %s", c.General.PfSenseXml, err)
					} else {
						err := c.PfSense.Finalize()
						if err != nil {
							logger.Errorf("cannot finalize pfSense configuration: %v", err)
						}
						c.PfSenseMode = true
						logger.Info("pfSense configuration valid: entering pfSense mode")
					}
				}
			}
		}
	} else {
		logger.Warn("no pfSense XML configuration file")
	}

	return changed, err
}

func (m Management) GetKey() inform.Key {
	if len(m.Key) == 0 {
		return inform.DefaultKey
	}
	keyBytes, err := inform.KeyFromString(m.Key)
	if err != nil {
		return inform.DefaultKey
	}
	return keyBytes
}

func (m Management) GetCryptoMode() int {
	if m.UseAesGcm {
		return inform.GCM
	}
	return inform.CBC
}

func New(path string, jsonFormat bool) (*Config, error) {
	var config Config
	config.path = path
	if jsonFormat {
		config.useJson = true
	}
	err := config.Read()
	config.setUpLog()
	changed, err := config.check()
	if changed {
		_ = config.Write()
	}
	return &config, err
}
