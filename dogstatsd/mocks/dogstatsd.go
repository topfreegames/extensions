// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package mocks

// ClientMock is a struct that implements dogstatsd.Client
type ClientMock struct {
	Counts map[string]int64
	Gauges map[string]float64
}

// NewClientMock ctor
func NewClientMock() *ClientMock {
	return &ClientMock{
		Counts: map[string]int64{},
		Gauges: map[string]float64{},
	}
}

// Incr adds 1 to Counts[name]
func (d *ClientMock) Incr(name string, tags []string, rate float64) error {
	d.Counts[name]++
	return nil
}

// Count adds value to Counts[name]
func (d *ClientMock) Count(name string, value int64, tags []string,
	rate float64) error {
	d.Counts[name] += value
	return nil
}

// Gauge sets value to Gauges[name]
func (d *ClientMock) Gauge(name string, value float64, tags []string,
	rate float64) error {
	d.Gauges[name] = value
	return nil
}
