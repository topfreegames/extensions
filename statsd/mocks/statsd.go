/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package mocks

//StatsDClientMock should be used for tests that need to send xmpp messages to StatsD
type StatsDClientMock struct {
	Counts  map[string]int
	Gauges  map[string]interface{}
	Timings map[string]interface{}
	Closed  bool
	Flushed bool
}

//NewStatsDClientMock creates a new instance
func NewStatsDClientMock() *StatsDClientMock {
	return &StatsDClientMock{
		Flushed: false,
		Closed:  false,
		Counts:  map[string]int{},
		Gauges:  map[string]interface{}{},
		Timings: map[string]interface{}{},
	}
}

//Increment stores the new count in a map
func (m *StatsDClientMock) Increment(bucket string) {
	m.Counts[bucket]++
}

//Count stores the new count in a map
func (m *StatsDClientMock) Count(bucket string, delta interface{}) {
	m.Counts[bucket] += delta.(int)
}

//Gauge stores the count in a map
func (m *StatsDClientMock) Gauge(bucket string, value interface{}) {
	m.Gauges[bucket] = value
}

//Timing stores the count in a map
func (m *StatsDClientMock) Timing(bucket string, value interface{}) {
	m.Timings[bucket] = value
}

//Flush records that it did flush
func (m *StatsDClientMock) Flush() {
	m.Flushed = true
}

//Close records that it is closed
func (m *StatsDClientMock) Close() {
	m.Closed = true
}
