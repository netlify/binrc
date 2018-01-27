# Binrc

Binrc is a command line application to manage different versions of binaries on GitHub releases.

Binrc doesn't try to be smart about version schemas and it only search for exact version matches.

Binrc is released under the [MIT License](LICENSE).
Please make sure you understand its [implications and guarantees](https://writing.kemitchell.com/2016/09/21/MIT-License-Line-by-Line.html).

## Installation

You can download prebuilt binaries from the [releases](https://github.com/netlify/binrc/releases).

You can also install binrc to install binrc:

```
binrc install netlify/binrc 0.2.0
```

Or you can download the cutting edge version with `go get`:

```
go get github.com/netlify/binrc
```

## Usage

There is a main subcommand, `install`.

### install subcommand

The `install` subcommand Installs a new binary. If that version is not already in the host's
cache, Binrc will try to fetch it from GitHub's releases and keep it in its cache.

```
binrc install spf13/hugo 0.19
```

### Versions as environment variables

Binrc supports settings binary version numbers as environment variables. It will search for the binary name followed by `_VERSION`
in the environment to configure the right version.

```
HUGO_VERSION=0.19 binrc install spf13/hugo
```

### Aliases

Binrc keeps a list of aliases to make installing binaries more easy. If a project name is not in `OWNER/NAME` for, Binrc will
check the list of aliases to try to resolve the project.

This is the current known list:

- hugo: spf13/hugo
- gutenberg: keats/gutenberg
