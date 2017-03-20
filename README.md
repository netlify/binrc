# Binrc

Binrc is a command line application to manage different versions of binaries on GitHub releases.

Binrc doesn't try to be smart about version schemas and it only search for exact version matches.

Binrc is released under the [MIT License](LICENSE).
Please make sure you understand its [implications and guarantees](https://writing.kemitchell.com/2016/09/21/MIT-License-Line-by-Line.html).

## Installation

You can download prebuilt binaries from the [releases](https://github.com/netlify/binrc/releases).

You can also use binrc to install binrc:

```
binrc use netlify/binrc 0.2.0
```

Or you can download the cutting edge version with `go get`:

```
go get github.com/netlify/binrc
```

## Usage

There are two main subcommands, `use` and `exec`. By default, Binrc executes `use` if there is no subcommand specified.

### Use subcommand

The `use` subcommand sets the version of the binary specified in the host PATH. If that version is not already in the host's
cache, Binrc will try to fetch it from GitHub's releases and keep it in its cache.

```
binrc use spf13/hugo 0.19
```

### Exec subcommand

The `exec` subcommand runs the given command with a specified version. This is a one-off operation that won't change the
system's PATH.

```
binrc exec --version 0.19 hugo build
```

### Versions as environment variables

Binrc supports settings binary version numbers as environment variables. It will search for the binary name followed by `_VERSION`
in the environment to configure the right version.

```
HUGO_VERSION=0.19 binrc exec hugo build
```

```
HUGO_VERSION=0.19 binrc use spf13/hugo
```
