// +build tools

package tools

// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// for more details

import (
	_ "github.com/rakyll/statik"
	_ "github.com/golang/lint/golint"
)
