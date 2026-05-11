package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testutils "github.com/mathwizz/testing/utils"
)

func TestWebServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Web-Server Suite")
}

var _ = testutils.AttachResourceReporter("../testing/reports")
