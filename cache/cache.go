package cache

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
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

	releasesTemplate = "https://api.github.com/repos/%s/releases"
	binTemplate      = "%s_%s_linux_amd64/%s_%s_linux_amd64"
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

// Represents the GitHub API for the releases from a repository.
type ReleaseAssets struct {
	Browser_download_url string
}
type Releases struct {
	Tag_name string
	Assets   []ReleaseAssets
}

// URL returns the URL to download the tarball from.
func (p *Project) URL() string {
	var tarballUrl string
	url := fmt.Sprintf(releasesTemplate, p.FullName)
	res, _ := http.Get(url)
	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()
	for {
		var releases []Releases
		if err := dec.Decode(&releases); err == io.EOF {
			break
		}
		for _, r := range releases {
			// Checks if the tag name exists
			if r.Tag_name == p.Version || r.Tag_name == p.cleanVersion {
				for _, a := range r.Assets {
					n := strings.ToLower(a.Browser_download_url)
					// Gets the correct url for the tarball
					if strings.Contains(n, "linux") && strings.Contains(n, "64bit") && strings.HasSuffix(n, ".tar.gz") {
						tarballUrl = a.Browser_download_url
					}
				}
			}
		}
	}
	return tarballUrl
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
		os.RemoveAll(parent)
		return errors.Wrapf(err, "error preparing directory to download %s", project.FullName)
	}
	defer os.RemoveAll(tmp)

	url := project.URL()
	res, err := http.Get(url)
	if err != nil {
		os.RemoveAll(parent)
		return errors.Wrapf(err, "error downloading project %s, from %s:", project.FullName, url)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		os.RemoveAll(parent)
		return errors.Errorf("error downloading project %s, from %s - binary doesn't seem to exist: %v", project.FullName, url, res.StatusCode)
	}

	if err := untar(res.Body, tmp); err != nil {
		os.RemoveAll(parent)
		return errors.Wrapf(err, "error unpacking file for %s, from %s", project.FullName, url)
	}

	bin := project.BinaryName()
	fp := path.Join(tmp, bin)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		fp = path.Join(tmp, project.Name)
	}
	if err := os.Rename(fp, destination); err != nil {
		os.RemoveAll(parent)
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
