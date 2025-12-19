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
	fmt.Println("  -d --deep        Clone without --depth=1 that useful for dev")
	fmt.Println("  -n --no-recurse  Do not update getdeps of the root modules")
	fmt.Println("  -e --enable      Comma-separated list of active profiles")
	fmt.Println("")
	fmt.Println("gitdeps " + Version)
}

type Options struct {
	force     bool
	update    bool
	deep      bool
	noRecurse bool
	noClean   bool
	profiles  []string
	usedProfs map[string]bool
}

func Execute(args []string) {
	opts := &Options{
		force:     false,
		update:    false,
		deep:      false,
		noRecurse: false,
		noClean:   false,
		profiles:  []string{},
		usedProfs: map[string]bool{},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-h" || arg == "--help" {
			PrintHelp()
			os.Exit(0)
		}

		if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'f':
					opts.force = true
				case 'u':
					opts.update = true
				case 'd':
					opts.deep = true
				case 'n':
					opts.noRecurse = true
				case 'c':
					opts.noClean = true
				case 'e':
					if i+1 >= len(args) {
						fmt.Println("Missing value for enable")
						fmt.Println("")
						PrintHelp()
						os.Exit(1)
					}
					i++
					opts.profiles = strings.Split(args[i], ",")
				default:
					fmt.Println("Unknown flag: -" + string(ch))
					fmt.Println("")
					PrintHelp()
					os.Exit(1)
				}
			}
			continue
		}

		switch arg {
		case "--force":
			opts.force = true
		case "--update":
			opts.update = true
		case "--deep":
			opts.deep = true
		case "--no-recurse":
			opts.noRecurse = true
		case "--no-clean":
			opts.noClean = true
		case "--enable":
			if i+1 >= len(args) {
				fmt.Println("Missing value for enable")
				fmt.Println("")
				PrintHelp()
				os.Exit(1)
			}
			i++
			opts.profiles = strings.Split(args[i], ",")
		default:
			fmt.Println("Unknown argument: " + arg)
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

		if len(module.Option) > 0 {
			active := false
			for _, p := range module.Option {
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
				fmt.Println("Skipped (exists) " + fullPath)
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

		{
			var args []string
			args = append(args, "fetch")
			if !opts.deep {
				args = append(args, "--depth", "1")
			}
			args = append(args, "origin")
			if module.Branch != "" {
				args = append(args, module.Branch)
			} else if module.Commit != "" {
				args = append(args, module.Commit)
			} else if module.Tag != "" {
				args = append(args, "tag", module.Tag)
			}
			err := RunCommand(fullPath, "git", args...)
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
			childOpts := &Options{
				force:     opts.force,
				update:    opts.update,
				noRecurse: opts.noRecurse,
				noClean:   opts.noClean,
				profiles:  module.Define,
				usedProfs: make(map[string]bool),
			}
			err = UpdateDeps(fullPath, childOpts)
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
	Option  []string `json:"option,omitempty"`
	Define  []string `json:"define,omitempty"`
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
