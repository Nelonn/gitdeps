package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) > 1 && (StrArrContains(os.Args[1:], "-h") || StrArrContains(os.Args[1:], "--help")) {
		PrintHelp()
		os.Exit(0)
	}

	Execute(os.Args[1:])
}

func PrintHelp() {
	fmt.Println("Usage: gitdeps [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --skip  Skip existing directories")
}

func Execute(args []string) {
	skip := false

	if len(args) == 1 {
		if args[0] == "--skip" {
			skip = true
		} else {
			fmt.Println("Unknown argument at position 0: " + args[0])
			fmt.Println("")
			PrintHelp()
			os.Exit(1)
		}
	}
	if len(args) > 2 {
		PrintHelp()
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = UpdateDeps(workingDir, skip)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Updated gitdeps")
	}
}

func UpdateDeps(workingDir string, skip bool) error {
	depsFile := path.Join(workingDir, "gitdeps.json")

	file, err := os.OpenFile(depsFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var modules ModuleMap

	err = json.Unmarshal(fileData, &modules)
	if err != nil {
		return err
	}

	// Verifying
	for modulePath, module := range modules {
		fullPath := filepath.Clean(filepath.Join(workingDir, modulePath))
		if !strings.HasPrefix(fullPath, workingDir) {
			return errors.New(depsFile + ": '" + modulePath + "': Attempting to access a path outside of the base directory")
		}
		if StrArrMoreThanOnePresent([]string{module.Branch, module.Commit, module.Tag}) {
			return errors.New(depsFile + ": '" + modulePath + "': You can specify only one of `branch`, `commit` or `tag`.")
		}
	}

	for modulePath, module := range modules {
		fullPath := filepath.Clean(filepath.Join(workingDir, modulePath))

		_, err := os.Lstat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				err := os.MkdirAll(fullPath, 0777)
				if err != nil {
					return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
				}
				// Continue execution
			} else {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		} else {
			if skip {
				continue
			}
			err := os.RemoveAll(fullPath)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
			err = os.MkdirAll(fullPath, 0777)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
			// Continue execution
		}

		err = RunCommand(fullPath, "git", "init")
		if err != nil {
			return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
		}

		err = RunCommand(fullPath, "git", "remote", "add", "origin", module.Remote)
		if err != nil {
			return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
		}

		if module.Branch != "" {
			err = RunCommand(fullPath, "git", "fetch", "origin", module.Branch)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		}
		if module.Commit != "" {
			err = RunCommand(fullPath, "git", "fetch", "origin", module.Commit)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		}
		if module.Tag != "" {
			err = RunCommand(fullPath, "git", "fetch", "origin", "tag", module.Tag)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		}

		err = RunCommand(fullPath, "git", "reset", "--hard", "FETCH_HEAD")
		if err != nil {
			return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
		}

		err = RunCommand(fullPath, "git", "submodule", "init")
		if err != nil {
			return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
		}

		err = RunCommand(fullPath, "git", "submodule", "update", "--recursive")
		if err != nil {
			return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
		}

		potentialSubdeps := path.Join(fullPath, "gitdeps.json")
		subInfo, err := os.Lstat(potentialSubdeps)
		if err != nil {
			continue
		}
		if !subInfo.Mode().IsRegular() {
			continue
		}
		err = UpdateDeps(fullPath, skip)
		if err != nil {
			return err
		}
	}

	return nil
}

type Module struct {
	Remote string `json:"remote"`
	Branch string `json:"branch,omitempty"`
	Commit string `json:"commit,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

type ModuleMap map[string]Module

func StrArrContains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func StrArrMoreThanOnePresent(arr []string) bool {
	present := false
	for _, s := range arr {
		if s != "" {
			if present {
				return true
			} else {
				present = true
			}
		}
	}
	return false
}

func RunCommand(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
