package jobless

import (
	"bytes"
	"log"
	"path/filepath"
	"text/template"
)

func ResolveVariables(
	variables *map[string]string,
	variableStore *ChainedVariableStore,
	path string,
) {

	for key, value := range *variables {
		(*variables)[key] = ResolveVariableString(
			value,
			variableStore,
			path,
		)

		log.Println("Resolved", key, "=", (*variables)[key])
	}
}

func ResolveVariableString(
	input string,
	variableStore *ChainedVariableStore,
	path string,
) string {
	functions := template.FuncMap{
		"path": func(inPath string) string {
			if filepath.IsAbs(inPath) {
				return inPath
			} else {
				return filepath.Join(path, inPath)
			}
		},
		"var": func(name string) string {
			if variableStore == nil {
				return name
			}
			if value, found := variableStore.Get(name); found {
				return value
			} else {
				log.Printf("Unknown variable \"%v\" referenced in %v", name, path)
				return name
			}
		},
	}

	buffer := bytes.NewBufferString("")

	t := template.New("variableTemplate")
	t.Funcs(functions)
	buffer.Reset()
	parsed, err := t.Parse(input)

	if err != nil {
		log.Printf(
			"Error parsing variable %v in %v: %v\n",
			input,
			path,
			err,
		)

		return input
	} else {
		parsed.Execute(buffer, nil)
		return buffer.String()
	}

}
