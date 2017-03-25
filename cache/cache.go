package cache

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	// DefaultStorePath is the name
	// of the default directory where
	// binaries and bundles are stored.
	DefaultStorePath = ".binrc"

	tarballTemplate = "https://github.com/%s/releases/download/%s/%s_%s_Linux-64bit.tar.gz"
	binTemplate     = "%s_%s_linux_amd64/%s_%s_linux_amd64"
)

// aliases is a map of known project aliases
// to make finding project more easy.
var aliases = map[string]string{
	"hugo": "spf13/hugo",
}

// Project represents a project managed with Binrc.
// It holds information about its version, cache and binary paths.
type Project struct {
	FullName     string
	Version      string
	Owner        string
	Name         string
	FullPath     string
	cleanVersion string
}

// URL returns the URL to download the tarball from.
func (p *Project) URL() string {
	return fmt.Sprintf(tarballTemplate, p.FullName, p.Version, p.Name, p.cleanVersion)
}

// BinaryName returns the name of the binary to look for
// after the tarball is extracted.
func (p *Project) BinaryName() string {
	return fmt.Sprintf(binTemplate, p.Name, p.cleanVersion, p.Name, p.cleanVersion)
}

func newProject(name, version string) (*Project, error) {
	name = strings.Trim(name, "/")
	if !strings.Contains(name, "/") {
		p, exist := aliases[name]
		if !exist {
			return nil, errors.Errorf("invalid project name %s. it should have a format like `netlify/binrc`.", name)
		}

		name = p
	}

	nwo := strings.SplitN(name, "/", 2)
	if version == "" {
		ev := os.Getenv(fmt.Sprintf("%s_VERSION", strings.ToUpper(nwo[1])))
		if ev == "" {
			return nil, errors.Errorf("unknown project version for %s", name)
		}
		version = ev
	}

	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return &Project{
		FullName:     name,
		Version:      version,
		Owner:        nwo[0],
		Name:         nwo[1],
		cleanVersion: strings.TrimLeft(version, "v"),
	}, nil
}

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
	project, err := newProject(name, version)
	if err != nil {
		return nil, err
	}

	p := c.binaryPath(project)
	_, err = exec.LookPath(p)
	if err != nil {
		if err := download(project, p); err != nil {
			return nil, err
		}

		project.FullPath = p
	} else {
		project.FullPath = p
	}

	return project, nil
}

func (c *Cache) binaryPath(project *Project) string {
	p := []string{c.storePath, "binaries", project.FullName, project.Version, project.Name}
	return path.Join(p...)
}

func download(project *Project, destination string) error {
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return errors.Wrapf(err, "error preparing cache directory for %s", project.FullName)
	}

	tmp, err := ioutil.TempDir(parent, "binrc-")
	if err != nil {
		return errors.Wrapf(err, "error preparing directory to download %s", project.FullName)
	}
	defer os.RemoveAll(tmp)

	url := project.URL()
	res, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "error downloading project %s, from %s:", project.FullName, url)
	}
	defer res.Body.Close()

	if err := untar(res.Body, tmp); err != nil {
		return errors.Wrapf(err, "error unpacking file for %s, from %s", project.FullName, url)
	}

	bin := project.BinaryName()
	fp := path.Join(tmp, bin)
	if err := os.Rename(fp, destination); err != nil {
		return errors.Wrapf(err, "error renaming %s to %s", fp, destination)
	}

	return nil
}

func untar(reader io.Reader, destination string) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(destination, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return errors.Wrapf(err, "error creating directory tree: %s", path)
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return errors.Wrapf(err, "error creating directory tree: %s", path)
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return errors.Wrapf(err, "error creating file: %s", path)
		}
		defer file.Close()

		_, err = io.Copy(file, tr)
		if err != nil {
			return errors.Wrapf(err, "error creating file: %s", file)
		}
	}

	return nil
}
