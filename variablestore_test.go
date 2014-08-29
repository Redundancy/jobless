package jobless

import (
	"testing"
)

func TestEmptyVariableStoreCannotResolveVariable(t *testing.T) {
	v := &ChainedVariableStore{
		Variables: map[string]string{},
	}

	_, err := v.Resolve("foo")

	if err == nil {
		t.Errorf("An empty variable store should not resolve a variable")
	}
}

func TestVariableStoreResolvesLocalSimpleValue(t *testing.T) {
	const EXPECTED = "bar"

	v := &ChainedVariableStore{
		Variables: map[string]string{
			"foo": EXPECTED,
		},
	}

	val, err := v.Resolve("foo")

	if val != EXPECTED {
		t.Errorf(
			"Did not get expected value \"%v\", instead got \"%v\"",
			EXPECTED, val)
	}
	if err != nil {
		t.Errorf("Resolve should not cause an error")
	}
}

func TestVariableStoreResolvesParentSimpleValue(t *testing.T) {
	const EXPECTED = "bar"

	parent := &ChainedVariableStore{
		Variables: map[string]string{
			"foo": EXPECTED,
		},
	}

	v := &ChainedVariableStore{
		Parent: parent,
	}

	val, err := v.Resolve("foo")

	if val != EXPECTED {
		t.Errorf(
			"Did not get expected value \"%v\", instead got \"%v\"",
			EXPECTED, val)
	}
	if err != nil {
		t.Errorf("Resolve should not cause an error")
	}
}

func TestVariableStoreResolvesComplexLocalValue(t *testing.T) {
	const EXPECTED = "bar"

	v := &ChainedVariableStore{
		Variables: map[string]string{
			"foo": EXPECTED,
			"zoo": "{{ var `foo` }}",
		},
	}

	val, err := v.Resolve("zoo")

	if val != EXPECTED {
		t.Errorf(
			"Did not get expected value \"%v\", instead got \"%v\"",
			EXPECTED, val)
	}
	if err != nil {
		t.Errorf("Resolve should not cause an error")
	}
}

/*
func TestSensibleResolveError(t *testing.T) {
	const EXPECTED = "bar"

	parent := &ChainedVariableStore{
		Path: "B",
		Variables: map[string]string{
			"foo": EXPECTED,
		},
	}

	v := &ChainedVariableStore{
		Parent: parent,
		Path:   "A",
		Variables: map[string]string{
			"zagbat": "{{ var `shoo` }}",
		},
	}

	_, err := v.Resolve("zagbat")

	t.Error(err)
}
*/
