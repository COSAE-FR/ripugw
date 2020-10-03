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
	"strconv"
)

type Noop struct {
	httpResponse
	ServerTime int `json:"server_time_in_utc"`
	Interval   int `json:"interval"`
}

func (msg *Noop) unmarshalMap(data map[string]interface{}) (err error) {
	serverTimeInterface, ok := data["server_time_in_utc"]
	if ok {
		serverTimeString, ok := serverTimeInterface.(string)
		if ok {
			serverTime, err := strconv.Atoi(serverTimeString)
			if err == nil {
				msg.ServerTime = serverTime
			}
		}
	}
	intervalInterface, ok := data["interval"]
	if ok {
		intervalFloat, ok := intervalInterface.(float64)
		if ok {
			msg.Interval = int(intervalFloat)
		}
	}
	return err
}

func (msg *Noop) Marshal() []byte {
	res, _ := json.Marshal(msg)
	return res
}

func (msg *Noop) String() string {
	return string(msg.Marshal())
}
