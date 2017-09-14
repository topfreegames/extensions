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
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alexcesaro/statsd"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/statsd/interfaces"
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
		var err error
		client, err = statsd.New(statsd.Address(host), statsd.FlushPeriod(flushPeriod), statsd.Prefix(prefix))

		if err != nil {
			l.WithError(err).Error("Error configuring statsd client.")
			return err
		}
	}

	s.Client = client
	l.Info("StatsD client configured")
	return nil
}

//Increment increments a metric in StatsD
func (s *StatsD) Increment(metric string) {
	s.Client.Increment(metric)
}

//Count increments a metric in StatsD by a delta
func (s *StatsD) Count(metric string, delta interface{}) {
	s.Client.Count(metric, delta)
}

//ReportGoStats reports go stats in statsd
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

func (s *StatsD) Flush() error {
	s.Client.Flush()
	return nil
}

//Cleanup closes statsd connection
func (s *StatsD) Cleanup() error {
	s.Client.Close()
	return nil
}
