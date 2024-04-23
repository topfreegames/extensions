// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package dogstatsd

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// Client is the interface to required dogstatsd functions
type Client interface {
	Incr(name string, tags []string, rate float64) error
	Count(name string, value int64, tags []string, rate float64) error
	Gauge(name string, value float64, tags []string, rate float64) error
	Timing(name string, value time.Duration, tags []string, rate float64) error
	Histogram(name string, value float64, tags []string, rate float64) error
	Distribution(name string, value float64, tags []string, rate float64) error
}

// DogStatsD is a wrapper to a dogstatsd.Client
type DogStatsD struct {
	Client Client
}

// New ctor
func New(host, prefix string) (*DogStatsD, error) {
	c, err := statsd.New(host, statsd.WithNamespace(prefix))
	if err != nil {
		return nil, err
	}
	d := &DogStatsD{Client: c}
	return d, nil
}

// NewFromClient ctor
func NewFromClient(C Client) *DogStatsD {
	return &DogStatsD{
		Client: C,
	}
}

// Incr calls Client.Incr
func (d *DogStatsD) Incr(name string, tags []string, rate float64) error {
	return d.Client.Incr(name, tags, rate)
}

// Count calls Client.Count
func (d *DogStatsD) Count(name string, value int64, tags []string,
	rate float64) error {
	return d.Client.Count(name, value, tags, rate)
}

// Gauge calls Client.Gauge
func (d *DogStatsD) Gauge(name string, value float64, tags []string,
	rate float64) error {
	return d.Client.Gauge(name, value, tags, rate)
}

// Timing calls Client.Timing
func (d *DogStatsD) Timing(name string, value time.Duration, tags []string,
	rate float64) error {
	return d.Client.Timing(name, value, tags, rate)
}

// Histogram calls Client.Histogram
func (d *DogStatsD) Histogram(name string, value float64, tags []string,
	rate float64) error {
	return d.Client.Histogram(name, value, tags, rate)
}

// Distribution calls Client.Distribution
func (d *DogStatsD) Distribution(name string, value float64, tags []string,
	rate float64) error {
	return d.Client.Distribution(name, value, tags, rate)
}
