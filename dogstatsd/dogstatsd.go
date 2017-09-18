// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package dogstatsd

import (
	"github.com/DataDog/datadog-go/statsd"
)

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

// New ctor
func New(host, prefix string) (*DogStatsD, error) {
	c, err := statsd.New(host)
	if err != nil {
		return nil, err
	}
	c.Namespace = prefix
	d := &DogStatsD{client: c}
	return d, nil
}

// NewFromClient ctor
func NewFromClient(client Client) *DogStatsD {
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
