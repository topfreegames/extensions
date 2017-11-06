// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package dogstatsd_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDogstatsd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dogstatsd Suite")
}

var (
	mockCtrl *gomock.Controller
)

var _ = BeforeEach(func() {
	mockCtrl = gomock.NewController(GinkgoT())
})
