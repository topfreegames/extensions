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
	"github.com/Sirupsen/logrus/hooks/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/statsd/mocks"
	"github.com/topfreegames/extensions/util"
)

var _ = Describe("StatsD Extension", func() {
	var config *viper.Viper
	var mockClient *mocks.StatsDClientMock
	logger, hook := test.NewNullLogger()
	BeforeEach(func() {
		var err error
		config, err = util.NewViperWithConfigFile("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())
		mockClient = mocks.NewStatsDClientMock()
		hook.Reset()
	})

	Describe("[Unit]", func() {
		Describe("Increment Metric", func() {
			It("should increment counter in statsd", func() {
				statsd, err := NewStatsD(config, logger, mockClient)
				Expect(err).NotTo(HaveOccurred())
				defer statsd.Cleanup()

				statsd.Increment("sent")
				statsd.Increment("sent")

				Expect(mockClient.Counts["sent"]).To(Equal(2))
			})
		})

		Describe("Count on metric", func() {
			It("should increment the metric by count", func() {
				statsd, err := NewStatsD(config, logger, mockClient)
				Expect(err).NotTo(HaveOccurred())
				defer statsd.Cleanup()

				statsd.Count("sent", 2)
				Expect(mockClient.Counts["sent"]).To(Equal(2))
			})
		})

		Describe("Flush call", func() {
			It("should call flush on statsd client", func() {
				statsd, err := NewStatsD(config, logger, mockClient)
				Expect(err).NotTo(HaveOccurred())
				defer statsd.Cleanup()

				Expect(mockClient.Flushed).To(Equal(false))
				statsd.Flush()
				Expect(mockClient.Flushed).To(Equal(true))
			})
		})
	})

	Describe("[Integration]", func() {
		Describe("Creating new client", func() {
			It("should return connected client", func() {
				statsd, err := NewStatsD(config, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(statsd).NotTo(BeNil())
				Expect(statsd.Client).NotTo(BeNil())
				defer statsd.Cleanup()

				Expect(statsd.Config).NotTo(BeNil())
				Expect(statsd.Logger).NotTo(BeNil())
			})
		})
	})
})
