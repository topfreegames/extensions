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
		mC *mocks.DogStatsDClientMock
		d  *dogstatsd.DogStatsD
	)

	BeforeEach(func() {
		mC = mocks.NewDogStatsDClientMock()
		d = dogstatsd.NewDogStatsD(mC)
	})

	It("Increment", func() {
		d.Increment("key", []string{}, 1)
		Expect(mC.Counts["key"]).To(Equal(int64(1)))
	})

	It("Count", func() {
		d.Count("key", 4, []string{}, 1)
		d.Count("key", 3, []string{}, 1)
		Expect(mC.Counts["key"]).To(Equal(int64(7)))
	})
})
