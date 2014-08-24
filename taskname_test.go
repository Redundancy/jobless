package jobless

import (
	"encoding/json"
	"gopkg.in/yaml.v1"
	"testing"
)

type testStruct struct {
	Name TaskName
}

func TestJsonMarshal(t *testing.T) {
	r, err := json.Marshal(
		&testStruct{
			Name: []string{"a", "b"},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(r) != `{"Name":"a.b"}` {
		t.Errorf("Unexpected marshalling result: %v", string(r))
	}
}

func TestYamlMarshal(t *testing.T) {
	r, err := yaml.Marshal(
		&testStruct{
			Name: []string{"a", "b"},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(r)[:9] != "name: a.b" {
		t.Errorf("Unexpected marshalling result: \"%v\" ", string(r))
	}
}

func TestYamlUnmarshal(t *testing.T) {
	const INPUT = `name: a.b`

	v := &testStruct{}
	err := yaml.Unmarshal([]byte(INPUT), v)

	if err != nil {
		t.Fatal(err)
	}

	if len(v.Name) != 2 {
		t.Fatalf("Incorrect name length: %v", v.Name)
	}

	if v.Name[0] != "a" {
		t.Errorf("Incorrect first path: %v", v.Name[0])
	}
	if v.Name[1] != "b" {
		t.Errorf("Incorrect second path: %v", v.Name[1])
	}
}

func TestMatchStars(t *testing.T) {
	const TASK = "A.B.unittest"
	const PATTERN = "*.*.*test"

	if !TaskNameFromString(TASK).Matches(PATTERN) {
		t.Errorf("%v did not match %v", TASK, PATTERN)
	}
}

func TestMatchRecursiveStars(t *testing.T) {
	const TASK = "A.B.unittest"
	const PATTERN = "**.*test"

	if !TaskNameFromString(TASK).Matches(PATTERN) {
		t.Errorf("%v did not match %v", TASK, PATTERN)
	}
}

func TestMatchRecursiveStarsBothEnds(t *testing.T) {
	const TASK = "A.B.unittest.foo"
	const PATTERN = "**.B.**"

	if !TaskNameFromString(TASK).Matches(PATTERN) {
		t.Errorf("%v did not match %v", TASK, PATTERN)
	}
}
