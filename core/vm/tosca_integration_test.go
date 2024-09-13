package vm

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/tracing"
)

func TestGetInterpreter_ProducesInterpretersBasedOnConfiguration(t *testing.T) {
	var (
		a         = &EVMInterpreter{}
		b         = &EVMInterpreter{}
		none      InterpreterFactory
		useA      = func(*EVM) Interpreter { return a }
		useB      = func(*EVM) Interpreter { return b }
		A         = func(i Interpreter) bool { return i == a }
		B         = func(i Interpreter) bool { return i == b }
		Fresh     = func(i Interpreter) bool { return i != nil && i != a && i != b }
		noTracing = false
		Tracing   = true
	)

	// Defines a complete "truth" table for the GetInterpreter function.
	tests := []struct {
		tracing               bool
		interpreter           InterpreterFactory
		interpreterForTracing InterpreterFactory
		want                  func(Interpreter) bool
	}{
		// tracing, interpreter, interpreterForTracing, want
		{noTracing, none, none, Fresh},
		{noTracing, none, useA, Fresh},
		{noTracing, none, useB, Fresh},

		{noTracing, useA, none, A},
		{noTracing, useA, useA, A},
		{noTracing, useA, useB, A},

		{noTracing, useB, none, B},
		{noTracing, useB, useA, B},
		{noTracing, useB, useB, B},

		{Tracing, none, none, Fresh},
		{Tracing, none, useA, A},
		{Tracing, none, useB, B},

		{Tracing, useA, none, A},
		{Tracing, useA, useA, A},
		{Tracing, useA, useB, B},

		{Tracing, useB, none, B},
		{Tracing, useB, useA, A},
		{Tracing, useB, useB, B},
	}

	for i, test := range tests {
		config := Config{
			Interpreter:           test.interpreter,
			InterpreterForTracing: test.interpreterForTracing,
		}
		if test.tracing {
			config.Tracer = &tracing.Hooks{}
		}
		evm := &EVM{Config: config}
		got := getInterpreter(evm)
		if !test.want(got) {
			t.Errorf("unexpected interpreter, case %d -  isA: %t, isB: %t, isFresh: %t", i, A(got), B(got), Fresh(got))
		}
	}
}
