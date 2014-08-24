package main

import (
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func WalkJoblessFiles(root string, fn func(*JoblessFile) error) {
	filepath.Walk(
		root,
		func(path string, info os.FileInfo, e error) error {
			if info.IsDir() || filepath.Ext(path) != ".jobless" {
				return nil
			}

			f, e := os.Open(path)
			if e != nil {
				log.Printf(
					"Error opening file %v: %v",
					path,
					e,
				)
				return nil
			}
			defer f.Close()

			b, e := ioutil.ReadAll(f)
			if e != nil {
				log.Printf(
					"Error reading file %v: %v",
					path,
					e,
				)
				return nil
			}

			var jobFile JoblessFile
			jobFile.Filepath = path

			e = yaml.Unmarshal(b, &jobFile)
			//e = json.Unmarshal(b, &jobFile)
			if e != nil {
				log.Printf(
					"Error parsing yaml file %v: %v",
					path,
					e,
				)
				return nil
			}

			return fn(&jobFile)
		},
	)

}
