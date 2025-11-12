package main

// This file implements the math problem solver.
// It uses the govaluate library to safely evaluate mathematical expressions.

import (
	"fmt"
	"strconv"

	"github.com/Knetic/govaluate"
)

// SolveMath evaluates a mathematical expression and returns the result.
// It supports basic arithmetic operations: +, -, *, /.
// Returns an error if the expression is invalid or cannot be evaluated.
func SolveMath(problem string) (int, error) {
	if problem == "" {
		return 0, fmt.Errorf("problem cannot be empty")
	}

	expression, err := govaluate.NewEvaluableExpression(problem)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %w", err)
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		return 0, fmt.Errorf("evaluation failed: %w", err)
	}

	switch v := result.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("result is not a number: %v", v)
		}
		return intVal, nil
	default:
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}
}
