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

	if originalWd, err := os.Getwd(); err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		defer os.Chdir(originalWd)
	}

	err = WalkJoblessFiles(
		cwd,
		func(jobFile *JoblessFile) error {
			parent := findParentStore(jobFile.Filepath, pathStores)
			taskDir := filepath.Dir(jobFile.Filepath)

			/*if parent != nil {
				log.Println("Chaining", jobFile.Filepath, "to", parent.Path)
			}*/

			thisStore := &jobless.ChainedVariableStore{
				Path:      jobFile.Filepath,
				Variables: jobFile.Variables,
				Parent:    parent,
			}

			pathStores[taskDir] = thisStore

			for _, task := range jobFile.Tasks {
				if !task.Name.Matches(pattern) {
					continue
				}

				log.Println("--------", task.Name, "--------")
				var err error

				resolvedEnvironment := make(map[string]string, len(task.Environment))

				for k, v := range task.Environment {
					//log.Println("Resolving Env:", k)
					resolvedEnvironment[k], err = thisStore.ResolveString(v)

					if err != nil {
						return fmt.Errorf(
							"Could not resolve environment variable %v for task %v: %v",
							k,
							task.Name,
							err,
						)
					}
				}

				//log.Println("Resolving Exe:", task.Command[0])
				executable, err := thisStore.ResolveString(task.Command[0])

				if err != nil {
					return err
				}

				arguments := task.Command[1:]

				if filepath.Ext(executable) == ".bat" {
					arguments = append(
						[]string{"/c", executable},
						arguments...,
					)
					executable = "cmd.exe"
				}

				wd, err := thisStore.ResolveString(task.CWD)

				if err != nil {
					return fmt.Errorf(
						"Could not resolve CWD for task %v - %v: %v",
						task.Name,
						task.CWD,
						err,
					)
				}

				if wd == "" {
					wd = taskDir
				}

				if err = os.Chdir(wd); err != nil {
					return err
				}

				// Set the CWD correctly before creating the command object

				cmd := exec.Command(executable, arguments...)
				cmd.Dir = wd
				log.Println("Executing:", executable, arguments)

				setEnvironment := make(map[string]bool, len(task.Environment))
				cmd.Env = make([]string, 0, len(os.Environ()))

				for _, e := range os.Environ() {
					kv := strings.Split(e, "=")
					k := kv[0]

					if joblessEnvValue, present := resolvedEnvironment[k]; present {
						log.Println("Env", k, "=", joblessEnvValue)
						cmd.Env = append(cmd.Env, k+"="+joblessEnvValue)
						setEnvironment[k] = true
					} else {
						//log.Println("Env inherit:", e)
						cmd.Env = append(cmd.Env, e)
					}
				}

				for k, v := range resolvedEnvironment {
					if !setEnvironment[k] {
						log.Println("Env", k, "=", v)
						cmd.Env = append(cmd.Env, k+"="+v)
					}
				}

				outPipe, _ := cmd.StdoutPipe()
				defer outPipe.Close()

				errPipe, _ := cmd.StderrPipe()
				defer errPipe.Close()

				go io.Copy(os.Stdout, outPipe)
				go io.Copy(os.Stderr, errPipe)
				log.Println("Working Directory =", cmd.Dir)

				err = cmd.Run()
				if err == nil {
					log.Println("-------- Done --------")
				} else {
					return err
				}
			}

			return nil
		},
	)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return
}
