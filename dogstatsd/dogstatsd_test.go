package dogstatsd_test

import (
	"github.com/topfreegames/extensions/dogstatsd"
	"github.com/topfreegames/extensions/dogstatsd/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DogStatsD", func() {
	It("test", func() {
		mC := mocks.NewDogStatsDClientMock()
		d := dogstatsd.NewDogStatsD(mC)
		d.Increment("key", []string{}, 1)
		Expect(mC.Counts["key"]).To(Equal(int64(1)))
	})
})
