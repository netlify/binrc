package cache

// aliases is a map of known project aliases
// to make finding project more easy.
var aliases = map[string]string{
	"hugo": "spf13/hugo",
}

// Project represents a project managed with Binrc.
// It holds information about its version, cache and binary paths.
type Project struct {
	Name       string
	Version    string
	BundlePath string
	BinaryPath string
}
