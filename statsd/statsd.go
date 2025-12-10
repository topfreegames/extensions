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

package statsd

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smira/go-statsd"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/statsd/interfaces"
)

// StatsD for sending metrics
type StatsD struct {
	Client interfaces.StatsDClient
	Config *viper.Viper
	Logger *logrus.Logger
}

// NewStatsD for creating a new StatsD instance
func NewStatsD(config *viper.Viper, logger *logrus.Logger, clientOrNil ...interfaces.StatsDClient) (*StatsD, error) {
	q := &StatsD{
		Config: config,
		Logger: logger,
	}
	var client interfaces.StatsDClient
	if len(clientOrNil) == 1 {
		client = clientOrNil[0]
	}
	err := q.configure(client)
	return q, err
}

func (s *StatsD) loadConfigurationDefaults() {
	s.Config.SetDefault("extensions.statsd.host", "localhost:8125")
	s.Config.SetDefault("extensions.statsd.prefix", "test")
	s.Config.SetDefault("extensions.statsd.flushIntervalMs", 5000)
}

func (s *StatsD) configure(client interfaces.StatsDClient) error {
	s.loadConfigurationDefaults()

	host := s.Config.GetString("extensions.statsd.host")
	prefix := s.Config.GetString("extensions.statsd.prefix")
	flushIntervalMs := s.Config.GetInt("extensions.statsd.flushIntervalMs")
	flushPeriod := time.Duration(flushIntervalMs) * time.Millisecond

	l := s.Logger.WithFields(logrus.Fields{
		"host":            host,
		"prefix":          prefix,
		"flushIntervalMs": flushIntervalMs,
	})

	if client == nil {
		// Create smira/go-statsd client
		smiraClient := statsd.NewClient(host,
			statsd.MetricPrefix(prefix),
			statsd.FlushInterval(flushPeriod),
			statsd.MaxPacketSize(1400), // Safe UDP packet size
		)

		// Wrap the smira client to match our interface
		client = &statsdClientAdapter{client: smiraClient}
	}

	s.Client = client
	l.Info("StatsD client configured")
	return nil
}

// statsdClientAdapter adapts smira/go-statsd to our interface
type statsdClientAdapter struct {
	client *statsd.Client
}

func (a *statsdClientAdapter) Increment(metric string) {
	a.client.Incr(metric, 1)
}

func (a *statsdClientAdapter) Count(metric string, delta interface{}) {
	var value int64
	switch v := delta.(type) {
	case int:
		value = int64(v)
	case int64:
		value = v
	case uint64:
		value = int64(v)
	case float64:
		value = int64(v)
	default:
		// Try to convert to int64
		value = int64(fmt.Sprintf("%v", v)[0] - '0')
	}
	a.client.Incr(metric, value)
}

func (a *statsdClientAdapter) Gauge(metric string, value interface{}) {
	var gaugeValue int64
	switch v := value.(type) {
	case int:
		gaugeValue = int64(v)
	case int64:
		gaugeValue = v
	case uint64:
		gaugeValue = int64(v)
	case float64:
		gaugeValue = int64(v)
	default:
		gaugeValue = 0
	}
	a.client.Gauge(metric, gaugeValue)
}

func (a *statsdClientAdapter) Timing(metric string, value interface{}) {
	var duration time.Duration
	switch v := value.(type) {
	case time.Duration:
		duration = v
	case int64:
		duration = time.Duration(v)
	case uint64:
		duration = time.Duration(v)
	case float64:
		duration = time.Duration(v)
	default:
		duration = 0
	}
	a.client.Timing(metric, int64(duration))
}

func (a *statsdClientAdapter) Flush() {
	// smira/go-statsd flushes automatically based on FlushInterval
	// This is a no-op for compatibility
}

func (a *statsdClientAdapter) Close() {
	a.client.Close()
}

// Increment increments a metric in StatsD
func (s *StatsD) Increment(metric string) {
	s.Client.Increment(metric)
}

// Count increments a metric in StatsD by a delta
func (s *StatsD) Count(metric string, delta interface{}) {
	s.Client.Count(metric, delta)
}

// ReportGoStats reports go stats in statsd
func (s *StatsD) ReportGoStats(
	numGoRoutines int,
	allocatedAndNotFreed, heapObjects, nextGCBytes, pauseGCNano uint64,
) {
	s.Client.Gauge("num_goroutine", numGoRoutines)
	s.Client.Gauge("allocated_not_freed", allocatedAndNotFreed)
	s.Client.Gauge("heap_objects", heapObjects)
	s.Client.Gauge("next_gc_bytes", nextGCBytes)
	s.Client.Timing("gc_pause_duration_ms", pauseGCNano/1000000)
}

// Flush calls Flush from statsd client
func (s *StatsD) Flush() error {
	s.Client.Flush()
	return nil
}

// Cleanup closes statsd connection
func (s *StatsD) Cleanup() error {
	s.Client.Close()
	return nil
}
