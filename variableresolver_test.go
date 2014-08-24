package jobless

import (
	"testing"
)

func TestPathResolver(t *testing.T) {
	vars := map[string]string{
		"A": `{{path "bar"}}`,
	}

	ResolveVariables(
		&vars,
		nil,
		"C:\\foo",
	)

	if vars["A"] != "C:\\foo\\bar" {
		t.Errorf(
			"Incorrect value: %v",
			vars["A"],
		)
	}
}

func TestVariableResolver(t *testing.T) {
	vars := map[string]string{
		"A": `hell{{var "bar"}}`,
	}

	vs := NewChainedVariableStore(nil)
	vs.Variables["bar"] = "o"

	ResolveVariables(
		&vars,
		vs,
		"C:\\foo",
	)

	if vars["A"] != "hello" {
		t.Errorf(
			"Incorrect value: %v",
			vars["A"],
		)
	}
}
