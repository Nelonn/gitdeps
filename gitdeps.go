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
	"strconv"
	"strings"
)

// Version is dynamically set by the toolchain
var Version = "DEV"

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
	fmt.Println("  -f --force       Remove, then clone existing modules")
	fmt.Println("  -u --update      Update existing modules")
	fmt.Println("  -n --no-recurse  Do not update getdeps of the root modules")
	fmt.Println("")
	fmt.Println("gitdeps " + Version)
}

func Execute(args []string) {
	force := false
	update := false
	noRecurse := false

	if duplicate := CheckStrDuplicates(args); duplicate != "" {
		fmt.Println("Got duplicate argument: " + duplicate)
		fmt.Println("")
		PrintHelp()
		os.Exit(1)
	}
	for pos, arg := range args {
		if arg == "-f" || arg == "--force" {
			if force {
				fmt.Println("Duplicate arguments: force")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			force = true
		} else if arg == "-u" || arg == "--update" {
			if update {
				fmt.Println("Duplicate arguments: update")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			update = true
		} else if arg == "-n" || arg == "--no-recurse" {
			if noRecurse {
				fmt.Println("Duplicate arguments: no-recurse")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			noRecurse = true
		} else {
			fmt.Println("Unknown argument at position " + strconv.Itoa(pos) + ": " + arg)
			fmt.Println("")
			PrintHelp()
			os.Exit(1)
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = UpdateDeps(workingDir, update, force, noRecurse)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Updated gitdeps")
	}
}

func UpdateDeps(workingDir string, update bool, force bool, noRecurse bool) error {
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

	// Verifying and cleaning
	for modulePath, module := range modules {
		if strings.HasPrefix(modulePath, "//") || strings.HasPrefix(modulePath, "#") {
			continue
		}

		fullPath := filepath.Clean(filepath.Join(workingDir, modulePath))
		if !strings.HasPrefix(fullPath, workingDir) {
			return errors.New(depsFile + ": '" + modulePath + "': Attempting to access a path outside of the base directory")
		}
		if StrArrMoreThanOneNotEmpty([]string{module.Branch, module.Commit, module.Tag}) {
			return errors.New(depsFile + ": '" + modulePath + "': You can specify only one of `branch`, `commit` or `tag`.")
		}
	}

	for modulePath, module := range modules {
		if strings.HasPrefix(modulePath, "//") || strings.HasPrefix(modulePath, "#") {
			continue
		}

		fullPath := filepath.Clean(filepath.Join(workingDir, modulePath))

		repoInitialized := false

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
			if force {
				err := os.RemoveAll(fullPath)
				if err != nil {
					return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
				}
				err = os.MkdirAll(fullPath, 0777)
				if err != nil {
					return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
				}
				// Continue execution
			} else if update {
				repoInitialized = true
				// Continue execution
			} else {
				fmt.Println("Skipped " + fullPath)
				continue
			}
		}

		if !repoInitialized {
			err = RunCommand(fullPath, "git", "init")
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		} else {
			_ = RunCommand(fullPath, "git", "remote", "remove", "origin")
		}

		err = RunCommand(fullPath, "git", "remote", "add", "origin", module.URL)
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

		if module.Patches != nil && len(module.Patches) > 0 {
			for i, patch := range module.Patches {
				absPatchPath := path.Join(workingDir, patch)
				fmt.Println("Applying patch " + absPatchPath)
				err = RunCommand(fullPath, "git", "apply", absPatchPath)
				if err != nil {
					return errors.New(depsFile + ": '" + modulePath + "' patches[" + strconv.Itoa(i) + "]: " + err.Error())
				}
			}
		}

		if noRecurse {
			continue
		}

		subDeps := path.Join(fullPath, "gitdeps.json")
		subDepsInfo, err := os.Lstat(subDeps)
		if err != nil || !subDepsInfo.Mode().IsRegular() {
			err = RunCommand(fullPath, "git", "submodule", "update", "--init", "--recursive")
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		} else {
			err = UpdateDeps(fullPath, update, force, noRecurse)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Module struct {
	URL     string   `json:"url"`
	Branch  string   `json:"branch,omitempty"`
	Commit  string   `json:"commit,omitempty"`
	Tag     string   `json:"tag,omitempty"`
	Patches []string `json:"patches,omitempty"`
}

type ModuleMap map[string]Module

func RunCommand(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func StrArrContains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func StrArrMoreThanOneNotEmpty(arr []string) bool {
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

func CheckStrDuplicates(arr []string) string {
	stringMap := make(map[string]bool)

	for _, str := range arr {
		if _, exists := stringMap[str]; exists {
			return str
		}
		stringMap[str] = true
	}

	return ""
}
