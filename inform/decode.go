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
)

func Unmarshal(data []byte) (Message, error) {

	var result map[string]interface{}
	erro := json.Unmarshal(data, &result)
	if erro != nil {
		return nil, erro
	}
	msgType := result["_type"]
	switch msgType {
	case "setparam":
		msg := &SetParam{
			httpResponse: ResponseFromHttpCode(200).(httpResponse),
		}
		err := msg.unmarshalMap(result)
		return msg, err
	case "noop":
		msg := &Noop{
			httpResponse: ResponseFromHttpCode(200).(httpResponse),
		}
		err := msg.unmarshalMap(result)
		return msg, err
	case "cmd":
		msg := &Cmd{
			httpResponse: ResponseFromHttpCode(200).(httpResponse),
		}
		err := msg.unmarshalMap(result)
		return msg, err
	default:
		msg := &Noop{
			httpResponse: ResponseFromHttpCode(200).(httpResponse),
		}
		err := msg.unmarshalMap(result)
		return msg, err
	}

	return nil, nil
}
