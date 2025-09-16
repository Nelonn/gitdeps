# gitdeps

Git dependencies manager

Often run into problems when using git submodule? Then use a simple and useful tool called `gitdeps`


## Try it right now!

Clone this repository as example gitdeps project

```shell
git clone https://github.com/Nelonn/gitdeps
cd gitdeps
```

Run the getdeps command in a convenient way:

- Build it yourself with `go` installed in your system:
```shell
go build gitdeps.go
./gitdeps or .\gitdeps.exe
```

- Or use [Homebrew](https://brew.sh/) for Linux or macOS:

```shell
brew tap Nelonn/tap
brew install gitdeps
gitdeps
```

If the command completes without errors then everything is ready! You can see the `helloworld` submodule in the directory!

This repository contain this `gitdeps.json`:

```json
{
  "helloworld": {
    "url": "https://github.com/go-training/helloworld",
    "branch": "master"
  }
}
```

## Features

- Easy to use!
- Will save you from the frequent `HEAD detached ...`
- Patches support!
- Does not clone standard submodules if there is an `gitdeps.json` in the repository
- Zero dependencies!


## gitdeps.json

This file manages external project dependencies. Each key is the relative local path where the module will be cloned.

Use prefix `//` or `#` to disable module, example: `//third_party/mbedtls`

Module declaration structure:

- `url`: URL of the remote Git repository.
- `branch`: Fetches the latest commit on this branch.
- `commit`: Checksum of commit that you want to use (usually SHA-1)
- `tag`: Fetches a specific tag.
- `patches`: An optional array of paths to patch files. Paths are relative to the gitdeps file. Patches are applied after cloning.
- `option`: An optional array of profile names (e.g., `["cpp17", "dev"]`). The dependency is cloned only if at least one of the listed profiles is specified using `--enable` or by parent gitdeps file. If empty or omitted, the dependency is always cloned.
- `define`: An optional array of profiles to pass down to this dependency's own gitdeps.json file. Profiles specified via command-line arguments only apply to the root configuration.

You can specify only one of `branch`, `commit` or `tag` in a single module!

Example:

```json
{
  "third_party/mbedtls": {
    "url": "https://github.com/Mbed-TLS/mbedtls",
    "tag": "v3.6.0"
  },
  "third_party/fmt": {
    "url": "https://github.com/fmtlib/fmt",
    "commit": "c4f6fa71357b223b0ab8ac29577c6228fde8853d",
    "patches": ["third_party/some_fmt.patch"],
    "option": ["cpp17"],
    "define": ["sub_profile"]
  }
}
```

Profile, provided in example can be enabled using `gitdeps --enable cpp17`


## License

Source code licensed under MIT License

```
The MIT License (MIT)

Copyright (c) 2024 Michael Neonov <two.nelonn at gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
