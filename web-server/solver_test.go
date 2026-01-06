package main

// This file contains unit tests for the SolveMath function.
// Tests pure logic without database or HTTP dependencies.

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SolveMath", func() {
	When("solving valid mathematical expressions", func() {
		DescribeTable("returns correct results",
			func(problem string, expected int) {
				Expect(SolveMath(problem)).Should(Equal(expected))
			},
			Entry("addition", "2+2", 4),
			Entry("subtraction", "10-3", 7),
			Entry("multiplication", "5*10", 50),
			Entry("division", "20/4", 5),
			Entry("complex expression", "25+75", 100),
			Entry("expression with parentheses", "(10+5)*2", 30),
			Entry("multiple operations", "100-50+25", 75),
		)
	})

	When("handling invalid or edge case inputs", func() {
		DescribeTable("returns appropriate errors",
			func(problem string, errorSubstring string) {
				_, err := SolveMath(problem)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(errorSubstring))
			},
			Entry("empty string", "", "cannot be empty"),
			Entry("invalid characters", "abc", "invalid expression"),
			Entry("incomplete expression", "5+", "invalid expression"),
			Entry("division by zero", "10/0", "evaluation failed"),
		)

		It("should handle negative results", func() {
			Expect(SolveMath("5-10")).Should(Equal(-5))
		})

		It("should handle zero result", func() {
			Expect(SolveMath("5-5")).Should(Equal(0))
		})
	})
})
