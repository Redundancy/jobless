package main

import (
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	// this should be in 'directory' order
	fileinfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	dirnames := make([]string, 0, len(fileinfos))
	filenames := make([]string, 0, len(fileinfos))

	for _, info := range fileinfos {
		name := info.Name()

		if info.IsDir() {
			dirnames = append(dirnames, name)
		} else if filepath.Ext(name) == ".jobless" {
			filenames = append(filenames, name)
		}
	}

	sort.Strings(filenames)
	sort.Strings(dirnames)

	return append(filenames, dirnames...), nil
}

func doubleSectionStringSort(s []string, last int) {
	if len(s) == 0 {
		return
	}

	sort.Strings(s[0 : last+1])

	if last+1 < len(s) {
		sort.Strings(s[last+1 : len(s)])
	}
}

type WalkFunc func(path string, info os.FileInfo, err error) error

// Intended to be similar to filepath.Walk
// However, we have a need to look at files in a directory
// before we potentially recurse
func walk(path string, info os.FileInfo, walkFn WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

func Walk(root string, walkFn WalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return walkFn(root, nil, err)
	}
	return walk(root, info, walkFn)
}

func WalkJoblessFiles(root string, fn func(*JoblessFile) error) error {
	return Walk(
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

			err := fn(&jobFile)

			return err
		},
	)

}
