// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package dogstatsd_test

import (
	"time"

	"github.com/topfreegames/extensions/v9/dogstatsd"
	"github.com/topfreegames/extensions/v9/dogstatsd/mocks"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("DogStatsD", func() {
	var (
		mC *mocks.MockClient
		d  *dogstatsd.DogStatsD
	)

	BeforeEach(func() {
		mC = mocks.NewMockClient(mockCtrl)
		d = dogstatsd.NewFromClient(mC)
	})

	Describe("[Unit]", func() {
		It("Incr", func() {
			mC.EXPECT().Incr("key", []string{}, float64(1))
			d.Incr("key", []string{}, 1)
		})

		It("Count", func() {
			mC.EXPECT().Count("key", int64(4), []string{}, float64(1))
			mC.EXPECT().Count("key", int64(3), []string{}, float64(1))
			d.Count("key", 4, []string{}, 1)
			d.Count("key", 3, []string{}, 1)
		})

		It("Gauge", func() {
			mC.EXPECT().Gauge("key", 86.5, []string{}, float64(1))
			d.Gauge("key", 86.5, []string{}, 1)
			mC.EXPECT().Gauge("key", 42.0, []string{}, float64(1))
			d.Gauge("key", 42.0, []string{}, 1)
		})

		It("Timing", func() {
			mC.EXPECT().Timing("key", 100*time.Millisecond,
				[]string{}, float64(1))
			d.Timing("key", 100*time.Millisecond, []string{}, 1)

			mC.EXPECT().Timing("key", 200*time.Millisecond,
				[]string{}, float64(1))
			d.Timing("key", 200*time.Millisecond, []string{}, 1)
		})

		It("Histogram", func() {
			mC.EXPECT().Histogram("key", float64(100), []string{}, float64(1))
			d.Histogram("key", 100, []string{}, 1)

			mC.EXPECT().Histogram("key", float64(200), []string{}, float64(1))
			d.Histogram("key", 200, []string{}, 1)
		})
	})
})
