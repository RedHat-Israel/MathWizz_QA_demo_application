package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testutils "github.com/mathwizz/testing/utils"
)

func TestHistoryWorker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "History-Worker Suite")
}

var _ = testutils.AttachResourceReporter("../testing/reports")
