/*
 * Copyright (c) 2017 ZAP Québec
 * Copyright (c) 2020 Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package inform

import (
	"encoding/json"
	"strconv"
	"strings"
)

type SetParam struct {
	httpResponse
	ManagementConfig `json:"mgmt_cfg"`
	ServerTime       int `json:"server_time_in_utc"`
}

func (msg *SetParam) unmarshalMap(data map[string]interface{}) (err error) {
	msg.ManagementConfig = make(ManagementConfig)
	for key, rawValue := range data {
		switch key {
		case "mgmt_cfg":
			if mgmt, ok := rawValue.(string); ok {
				if err = msg.ManagementConfig.unmarshalStr(mgmt); err != nil {
					return err
				}
			}
		case "server_time_in_utc":
			if mgmt, ok := rawValue.(string); ok {
				msg.ServerTime, err = strconv.Atoi(mgmt)
				if err != nil {
					return err
				}
			} else if mgmt, ok := rawValue.(int); ok {
				msg.ServerTime = mgmt
			}
		}
	}
	return nil
}

func (msg *SetParam) Marshal() []byte {
	res, _ := json.Marshal(msg)
	return res
}

func (msg *SetParam) String() string {
	return string(msg.Marshal())
}

type ManagementConfig map[string]string

func (m ManagementConfig) unmarshalStr(str string) error {
	for _, line := range strings.Split(str, "\n") {
		i := strings.IndexByte(line, '=')
		if i != -1 {
			m[line[0:i]] = line[i+1:]
		}
	}
	return nil
}

func (m ManagementConfig) MarshalJSON() ([]byte, error) {
	str := ""
	for k, v := range m {
		str = str + k + "=" + v + "\n"
	}
	return json.Marshal(map[string]string(m))
}
