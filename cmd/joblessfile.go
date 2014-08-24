package main

import (
	"github.com/Redundancy/jobless"
)

type Task struct {
	// Format A.B.C
	Name jobless.TaskName `yaml:"Name"`

	// Is the current working directory different than
	// that of the file?
	// evaluated for variables
	CWD string `json:",omitempty" yaml:"CWD,omitempty"`

	// The command to execute
	Command []string `yaml:"Command,omitempty"`

	// Environment variables
	Environment map[string]string `json:",omitempty" yaml:",omitempty"`
}

type JoblessFile struct {
	// Where did this file come from?
	Filepath string `json:"-" yaml:"-"`

	// top level variables. Will be inherited by tasks here
	// and those in subdirectories.
	// will be evaluated for variables
	Variables map[string]string `json:",omitempty" yaml:"Variables,omitempty"`

	// List of named tasks that can be run
	Tasks []Task `json:",omitempty" yaml:"Tasks,omitempty"`
}
