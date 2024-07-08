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
    "remote": "https://github.com/go-training/helloworld",
    "branch": "master"
  }
}
```

## Features

- Easy to use!
- Will save you from the frequent `HEAD detached ...`
- Does not clone standard submodules if there is an `gitdeps.json` in the repository
- Zero dependencies!


## gitdeps.json

Key of every module means relative path to module

Use prefix `//` or `#` to disable module, example `//third_party/mbedtls`

- `remote`: URL to remote repository
- `branch`: Fetch latest commit on the branch
- `commit`: SHA-1 of commit that you want to use
- `tag`: Fetch specified tag

You can specify only one of `branch`, `commit` or `tag` in a single module!

Real world example:

```json
{
  "third_party/mbedtls": {
    "remote": "https://github.com/Mbed-TLS/mbedtls",
    "tag": "v3.6.0"
  },
  "third_party/fmt": {
    "remote": "https://github.com/fmtlib/fmt",
    "commit": "c4f6fa71357b223b0ab8ac29577c6228fde8853d"
  }
}
```


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
