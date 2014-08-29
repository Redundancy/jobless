package jobless

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ChainedVariableStore struct {
	Parent    *ChainedVariableStore
	Path      string
	Variables map[string]string
}

type ResolveErrors []error

func (r ResolveErrors) Error() string {
	result := make([]string, len(r))
	for i, v := range r {
		result[i] = v.Error()
	}
	return strings.Join(result, "\n")
}

func (c *ChainedVariableStore) Resolve(variable string) (string, error) {
	if value, present := c.Variables[variable]; present {
		if contains_variable(value) {
			// potentially a complex value
			val, err := resolve_complex(c, value)
			if err == nil {
				return val, err
			} else {
				return "", fmt.Errorf(
					"Could not resolve variable \"%v\":\"%v\" in %v: %v",
					variable,
					value,
					c.Path,
					err,
				)
			}
		} else {
			// avoid doing template parsing etc
			return value, nil
		}

	} else if c.Parent != nil {
		return c.Parent.Resolve(variable)
	}

	return "", fmt.Errorf("variable \"%v\" was not found", variable)
}

func (c *ChainedVariableStore) ResolveString(value string) (string, error) {
	if contains_variable(value) {
		// potentially a complex value
		val, err := resolve_complex(c, value)
		if err == nil {
			return val, err
		} else {
			return "", fmt.Errorf(
				"Could not resolve \"%v\" in %v: %v",
				value,
				c.Path,
				err,
			)
		}
	} else {
		// avoid doing template parsing etc
		return value, nil
	}
}

func contains_variable(s string) bool {
	return strings.Contains(s, "{{")
}

func resolve_complex(store *ChainedVariableStore, value string) (string, error) {

	resolveErrors := make(ResolveErrors, 0, 10)

	// Rather than get errors out of Execute(),
	// We collect errors directly from the function invocations
	// This avoids the issue of them getting wrapped by the template system
	functions := template.FuncMap{
		"path": func(inPath string) string {
			if filepath.IsAbs(inPath) {
				return inPath
			} else {
				return filepath.Join(filepath.Dir(store.Path), inPath)
			}
		},
		"var": func(name string) string {
			result, err := store.Resolve(name)
			if err != nil {
				resolveErrors = append(resolveErrors, err)
			} else {
				log.Println("Resolved", name, "=", result)
			}
			return result
		},
		"env": func(name string) string {
			return os.Getenv(name)
		},
	}

	buffer := bytes.NewBufferString("")

	t := template.New("")
	t.Funcs(functions)
	buffer.Reset()
	parsed, err := t.Parse(value)

	if err != nil {
		return "", err
	}
	err = parsed.Execute(buffer, nil)

	if err != nil {
		return "", err
	} else if len(resolveErrors) > 0 {
		return "", resolveErrors
	}
	return buffer.String(), nil

}
