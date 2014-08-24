package jobless

import (
	"testing"
)

func TestVariableStoreReturnsLocalValue(t *testing.T) {
	const VALUE = "b"
	const KEY = "a"

	vs := NewChainedVariableStore(nil)
	vs.Variables[KEY] = VALUE

	if val, has := vs.Get(KEY); !has {
		t.Error("Variable Store does not have the local variable")
	} else if val != VALUE {
		t.Errorf("Unexpected value %v", val)
	}
}

func TestVariableStoreReturnsChainedValue(t *testing.T) {
	const KEY = "a"
	const VALUE = "b"

	parent := NewChainedVariableStore(nil)
	parent.Variables[KEY] = VALUE

	child := NewChainedVariableStore(parent)

	if val, has := child.Get(KEY); !has {
		t.Error("Variable Store does not have the local variable")
	} else if val != VALUE {
		t.Errorf("Unexpected value %v", val)
	}
}

func TestOverrideHidesChainedValue(t *testing.T) {
	const KEY = "a"
	const VALUE = "b"
	const OVERRIDE = "c"

	parent := NewChainedVariableStore(nil)
	parent.Variables[KEY] = VALUE

	child := NewChainedVariableStore(parent)
	child.Variables[KEY] = OVERRIDE

	if val, has := child.Get(KEY); !has {
		t.Error("Variable Store does not have the local variable")
	} else if val != OVERRIDE {
		t.Errorf("Unexpected value %v", val)
	}
}
