package dogstatsd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDogstatsd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dogstatsd Suite")
}
