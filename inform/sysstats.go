/*
 * Copyright (c) 2017 ZAP Québec
 * Copyright (c) 2020 Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package inform

type SysStats struct {
	LoadAvg1  float64 `json:"loadavg_1"`
	LoadAvg5  float64 `json:"loadavg_5"`
	LoadAvg15 float64 `json:"loadavg_15"`
	MemBuffer uint64  `json:"mem_buffer"`
	MemTotal  uint64  `json:"mem_total"`
	MemUsed   uint64  `json:"mem_used"`
	Mem       uint64  `json:"mem"`
	Cpu       uint64  `json:"cpu"`
}
