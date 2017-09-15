// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package dogstatsd_test

import (
	"github.com/topfreegames/extensions/dogstatsd"
	"github.com/topfreegames/extensions/dogstatsd/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DogStatsD", func() {
	var (
		mC *mocks.ClientMock
		d  *dogstatsd.DogStatsD
	)

	BeforeEach(func() {
		mC = mocks.NewClientMock()
		d = dogstatsd.NewDogStatsD(mC)
	})

	It("Incr", func() {
		d.Incr("key", []string{}, 1)
		Expect(mC.Counts["key"]).To(Equal(int64(1)))
	})

	It("Count", func() {
		d.Count("key", 4, []string{}, 1)
		d.Count("key", 3, []string{}, 1)
		Expect(mC.Counts["key"]).To(Equal(int64(7)))
	})

	It("Gauge", func() {
		d.Gauge("key", 86.5, []string{}, 1)
		Expect(mC.Gauges["key"]).To(Equal(float64(86.5)))
		d.Gauge("key", 42.0, []string{}, 1)
		Expect(mC.Gauges["key"]).To(Equal(float64(42.0)))
	})
})
