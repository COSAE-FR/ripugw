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
	"fmt"
	"strconv"
)

type InformResponse interface {
	Message
	IsSuccess() bool
	HttpCode() int
}

type httpResponse struct {
	code int
}

func ResponseFromHttpCode(code int) InformResponse {
	return httpResponse{code}
}

func (r httpResponse) IsSuccess() bool {
	return r.code == 200
}

func (r httpResponse) HttpCode() int {
	return r.code
}

func (r httpResponse) Marshal() []byte {
	return []byte(r.String())
}

func (r httpResponse) String() string {
	return fmt.Sprintf(`{"code":%d}`, r.code)
}

func ParseString(value interface{}) (string, error) {
	result, ok := value.(string)
	if ok {
		return result, nil
	}
	return result, fmt.Errorf("not a string")
}

func ParseFloat(value interface{}) (float64, error) {
	result, ok := value.(float64)
	if ok {
		return result, nil
	}
	return result, fmt.Errorf("not a float64")
}

func ParseInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert value")
	}
}

func ParseBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case int:
		return v > 0, nil
	case float64:
		return int(v) > 0, nil
	case string:
		if v == "true" || v == "True" || v == "1" {
			return true, nil
		} else if v == "false" || v == "False" || v == "0" {
			return false, nil
		}
	case bool:
		return v, nil
	}
	return false, fmt.Errorf("cannot convert value")
}
