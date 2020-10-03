/*
 * Copyright (c) 2020 Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package inform

import (
	"encoding/json"
)

type Cmd struct {
	httpResponse
	ServerTime int    `json:"server_time_in_utc"`
	Command    string `json:"cmd"`
	UseAlert   bool   `json:"use_alert"`
	DeviceId   string `json:"device_id"`
	Time       int    `json:"time"`
	CmdId      string `json:"_id"`
}

func (msg *Cmd) unmarshalMap(data map[string]interface{}) (err error) {
	for key, rawValue := range data {
		switch key {
		case "server_time_in_utc":
			msg.ServerTime, _ = ParseInt(rawValue)
			continue
		case "cmd":
			msg.Command, _ = ParseString(rawValue)
			continue
		case "use_alert":
			msg.UseAlert, _ = ParseBool(rawValue)
			continue
		case "time":
			msg.Time, _ = ParseInt(rawValue)
			continue
		case "_id":
			msg.CmdId, _ = ParseString(rawValue)
			continue
		case "device_id":
			msg.DeviceId, _ = ParseString(rawValue)
			continue
		}
	}
	return nil
}

func (msg *Cmd) Marshal() []byte {
	res, _ := json.Marshal(msg)
	return res
}

func (msg *Cmd) String() string {
	return string(msg.Marshal())
}
