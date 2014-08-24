package main

import (
	"fmt"
	"github.com/Redundancy/jobless"
	"github.com/codegangsta/cli"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var app *cli.App = cli.NewApp()

func main() {
	app.Name = "jobless"
	app.Usage = "Organize your project commands effortlessly, then automate it."
	app.Flags = []cli.Flag{}

	runtime.GOMAXPROCS(runtime.NumCPU())

	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:        "find",
			ShortName:   "f",
			Usage:       "jobless find <pattern>",
			Description: `finds tasks with a given pattern. Defaults to ** (all)`,
			Action:      find,
			Flags:       []cli.Flag{},
		},
	)

	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:        "run",
			ShortName:   "r",
			Usage:       "jobless run <pattern>",
			Description: `runs tasks with a given pattern. Defaults to ** (all)`,
			Action:      run,
			Flags:       []cli.Flag{},
		},
	)

	app.Run(os.Args)
}

func findParentStore(
	path string,
	stores map[string]*jobless.ChainedVariableStore,
) *jobless.ChainedVariableStore {

	d := filepath.Dir(path)
	for d[len(d)-1] != filepath.Separator {
		if store, found := stores[d]; found {
			return store
		}
		d = filepath.Dir(d)
	}

	return nil
}

func find(c *cli.Context) {
	pattern := "**"
	if len(c.Args()) == 1 {
		pattern = c.Args()[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get cwd: %v", err)
		os.Exit(1)
	}

	WalkJoblessFiles(
		cwd,
		func(jobFile *JoblessFile) error {
			for _, task := range jobFile.Tasks {
				if task.Name.Matches(pattern) {
					fmt.Printf("%v in %v\n", task.Name, jobFile.Filepath)
				}
			}
			return nil
		},
	)

	return
}

func run(c *cli.Context) {
	pattern := "**"
	if len(c.Args()) == 1 {
		pattern = c.Args()[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get cwd: %v", err)
		os.Exit(1)
	}

	pathStores := make(map[string]*jobless.ChainedVariableStore)

	WalkJoblessFiles(
		cwd,
		func(jobFile *JoblessFile) error {
			parent := findParentStore(jobFile.Filepath, pathStores)
			taskDir := filepath.Dir(jobFile.Filepath)

			jobless.ResolveVariables(
				&jobFile.Variables,
				parent,
				taskDir,
			)

			thisStore := jobless.NewChainedVariableStore(parent)
			pathStores[taskDir] = thisStore
			thisStore.Variables = jobFile.Variables

			for _, task := range jobFile.Tasks {
				if !task.Name.Matches(pattern) {
					continue
				}

				log.Println("--------", task.Name, "--------")

				for k, v := range task.Environment {
					task.Environment[k] = jobless.ResolveVariableString(
						v,
						thisStore,
						taskDir,
					)
				}

				executable := jobless.ResolveVariableString(
					task.Command[0],
					thisStore,
					taskDir,
				)

				arguments := task.Command[1:]

				if filepath.Ext(executable) == ".bat" {
					arguments = append(
						[]string{"/c", executable},
						arguments...,
					)
					executable = "cmd"
				}

				cmd := exec.Command(executable, arguments...)

				setEnvironment := make(map[string]bool, len(task.Environment))
				cmd.Env = make([]string, 0, len(os.Environ()))

				for _, e := range os.Environ() {
					kv := strings.Split(e, "=")
					k := kv[0]

					if joblessEnvValue, present := task.Environment[k]; present {
						cmd.Env = append(cmd.Env, k+"="+joblessEnvValue)
						setEnvironment[k] = true
					} else {
						cmd.Env = append(cmd.Env, e)
					}
				}

				for k, v := range task.Environment {
					if !setEnvironment[k] {
						cmd.Env = append(cmd.Env, k+"="+v)
					}
				}

				outPipe, _ := cmd.StdoutPipe()
				defer outPipe.Close()

				errPipe, _ := cmd.StderrPipe()
				defer errPipe.Close()

				go io.Copy(os.Stdout, outPipe)
				go io.Copy(os.Stderr, errPipe)

				cmd.Dir = jobless.ResolveVariableString(
					task.CWD,
					thisStore,
					taskDir,
				)

				if cmd.Dir == "" {
					cmd.Dir = taskDir
				}

				err := cmd.Run()

				if err != nil {
					log.Println(err)
					break
				}

			}

			return nil
		},
	)

	return
}
