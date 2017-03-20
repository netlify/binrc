package cache

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// DefaultStorePath is the name
// of the default directory where
// binaries and bundles are stored.
const DefaultStorePath = ".binrc"

// Cache stores and retrieves projects
// from the system's cache and GitHub releases
type Cache struct {
	storePath string
}

// New initializes the cache with
// a given store path.
func New(storePath string) *Cache {
	return &Cache{storePath}
}

// GetOrSet fetches a project from the cache.
// If the project is not in the cache, it tries
// to donwload it from GitHub releases.
func (c *Cache) GetOrSet(name, version string) (*Project, error) {
	name = strings.Trim(name, "/")

	if !strings.Contains(name, "/") {
		p, exist := aliases[name]
		if !exist {
			return nil, errors.Errorf("invalid project name %s. it should have a format like `netlify/binrc`.")
		}

		name = p
	}

	project := &Project{
		Name:    name,
		Version: version,
	}

	p := c.binaryPath(name, version)
	_, err := exec.LookPath(p)
	if err != nil {
		// check bundles and download
		bp := c.bundlePath(name, version)
		fi, err := os.Stat(bp)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrapf(err, "bundle file read failed: %s", bp)
			}

			// download from GitHub
		}

		if fi.IsDir() {
			return nil, errors.Errorf("file is a directory: %s", bp)
		}

		// untar

	} else {
		project.BinaryPath = p
	}

	return project, nil
}

func (c *Cache) binaryPath(project, version string) string {
	nameWithOwner := strings.Split(project, "/")

	p := []string{c.storePath, "binaries"}
	p = append(p, nameWithOwner...)
	p = append(p, version, nameWithOwner[1])
	return path.Join(p...)
}

func (c *Cache) bundlePath(project, version string) string {
	nameWithOwner := strings.Split(project, "/")
	cleanVersion := strings.TrimLeft(version, "v")

	p := []string{c.storePath, "bundles"}
	p = append(p, nameWithOwner...)
	p = append(p, fmt.Sprintf("%s_%s_Linux-64bit.tar.gz", nameWithOwner[1], cleanVersion))
	return path.Join(p...)
}
