package noaa_test

import (
	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/gomega"

	"testing"
)

func TestNoaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Noaa Suite")
}
