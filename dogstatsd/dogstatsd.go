// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package dogstatsd

// Client is the interface to required dogstatsd functions
type Client interface {
	Incr(name string, tags []string, rate float64) error
	Count(name string, value int64, tags []string, rate float64) error
	Gauge(name string, value float64, tags []string, rate float64) error
}

// DogStatsD is a wrapper to a dogstatsd.Client
type DogStatsD struct {
	client Client
}

// NewDogStatsD ctor
func NewDogStatsD(client Client) *DogStatsD {
	return &DogStatsD{
		client: client,
	}
}

// Incr calls Client.Incr
func (d *DogStatsD) Incr(name string, tags []string, rate float64) error {
	return d.client.Incr(name, tags, rate)
}

// Count calls Client.Count
func (d *DogStatsD) Count(name string, value int64, tags []string,
	rate float64) error {
	return d.client.Count(name, value, tags, rate)
}

// Gauge calls Client.Gauge
func (d *DogStatsD) Gauge(name string, value float64, tags []string,
	rate float64) error {
	return d.client.Gauge(name, value, tags, rate)
}
