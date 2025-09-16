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
	fmt.Println("  -p --profiles    Comma-separated list of active profiles")
	fmt.Println("")
	fmt.Println("gitdeps " + Version)
}

type Options struct {
	force     bool
	update    bool
	noRecurse bool
	noClean   bool
	profiles  []string
	usedProfs map[string]bool
}

func Execute(args []string) {
	opts := &Options{
		force:     false,
		update:    false,
		noRecurse: false,
		noClean:   false,
		profiles:  []string{},
		usedProfs: map[string]bool{},
	}

	skipNext := false
	for pos, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-h" || arg == "--help" {
			PrintHelp()
			os.Exit(0)
		}
		if arg == "-f" || arg == "--force" {
			if opts.force {
				fmt.Println("Duplicate arguments: force")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			opts.force = true
		} else if arg == "-u" || arg == "--update" {
			if opts.update {
				fmt.Println("Duplicate arguments: update")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			opts.update = true
		} else if arg == "-n" || arg == "--no-recurse" {
			if opts.noRecurse {
				fmt.Println("Duplicate arguments: no-recurse")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			opts.noRecurse = true
		} else if arg == "-c" || arg == "--no-clean" {
			if opts.noClean {
				fmt.Println("Duplicate arguments: no-clean")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			opts.noClean = true
		} else if arg == "-p" || arg == "--profiles" {
			if len(opts.profiles) > 0 {
				fmt.Println("Duplicate arguments: profiles")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			if pos+1 >= len(args) {
				fmt.Println("Missing value for profiles")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			opts.profiles = strings.Split(args[pos+1], ",")
			skipNext = true
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

	err = UpdateDeps(workingDir, opts)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Updated gitdeps")
	}

	if len(opts.profiles) > 0 {
		for _, profile := range opts.profiles {
			if !opts.usedProfs[profile] {
				fmt.Println("Warning: profile not used:", profile)
			}
		}
	}
}

func UpdateDeps(workingDir string, opts *Options) error {
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

		if len(module.Profiles) > 0 {
			active := false
			for _, p := range module.Profiles {
				if StrArrContains(opts.profiles, p) {
					active = true
					opts.usedProfs[p] = true
				}
			}
			if !active {
				fmt.Println("Skipped (profile disabled) " + modulePath)
				continue
			}
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
			if opts.force {
				err := os.RemoveAll(fullPath)
				if err != nil {
					return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
				}
				err = os.MkdirAll(fullPath, 0777)
				if err != nil {
					return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
				}
				// Continue execution
			} else if opts.update {
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
			err = RunCommand(fullPath, "git", "fetch", "--depth", "1", "origin", module.Branch)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		}
		if module.Commit != "" {
			err = RunCommand(fullPath, "git", "fetch", "--depth", "1", "origin", module.Commit)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		}
		if module.Tag != "" {
			err = RunCommand(fullPath, "git", "fetch", "--depth", "1", "origin", "tag", module.Tag)
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		}

		err = RunCommand(fullPath, "git", "reset", "--hard", "FETCH_HEAD")
		if err != nil {
			return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
		}

		if !opts.noClean {
			err = RunCommand(fullPath, "git", "clean", "-fd")
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
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

		if opts.noRecurse {
			continue
		}

		subDeps := path.Join(fullPath, "gitdeps.json")
		subDepsInfo, err := os.Lstat(subDeps)
		if err != nil || !subDepsInfo.Mode().IsRegular() {
			err = RunCommand(fullPath, "git", "submodule", "update", "--init", "--recursive", "--depth", "1")
			if err != nil {
				return errors.New(depsFile + ": '" + modulePath + "': " + err.Error())
			}
		} else {
			err = UpdateDeps(fullPath, opts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Module struct {
	URL      string   `json:"url"`
	Branch   string   `json:"branch,omitempty"`
	Commit   string   `json:"commit,omitempty"`
	Tag      string   `json:"tag,omitempty"`
	Patches  []string `json:"patches,omitempty"`
	Profiles []string `json:"profiles,omitempty"`
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
