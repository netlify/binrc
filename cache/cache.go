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

	"github.com/BurntSushi/toml"
	version "github.com/hashicorp/go-version"
	_ "github.com/netlify/binrc/statik" // this is required
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
)

const (
	// DefaultStorePath is the name
	// of the default directory where
	// binaries and bundles are stored.
	DefaultStorePath = ".binrc"

	tarballTemplate = "https://github.com/%s/releases/download/%s/%s"
)

// aliases is a map of known project aliases
// to make finding project more easy.
var aliases = map[string]string{
	"hugo": "spf13/hugo",
}

type template struct {
	Range   string
	Tarball string
	Bin     string
}

type templates map[string][]template

var (
	defaultTemplate = &template{
		Tarball: "%s_v%s_Linux-64bit.tar.gz",
		Bin:     "%s_%s_linux_amd64/%s_%s_linux_amd64",
	}
)

// Cache stores and retrieves projects
// from the system's cache and GitHub releases
type Cache struct {
	templates templates
	storePath string
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
	template     *template
}

// URL returns the URL to download the tarball from.
func (p *Project) URL() string {
	tarballName := fmt.Sprintf(p.template.Tarball, p.Name, p.cleanVersion)
	return fmt.Sprintf(tarballTemplate, p.FullName, p.Version, tarballName)
}

// BinaryName returns the name of the binary to look for
// after the tarball is extracted.
func (p *Project) BinaryName() string {
	if strings.Contains(p.template.Bin, "%") {
		return fmt.Sprintf(p.template.Bin, p.Name, p.cleanVersion, p.Name, p.cleanVersion)
	}
	return p.template.Bin
}

func (c *Cache) newProject(name, versionString string) (*Project, error) {
	name = strings.Trim(name, "/")
	if !strings.Contains(name, "/") {
		p, exist := aliases[name]
		if !exist {
			return nil, errors.Errorf("invalid project name %s. it should have a format like `netlify/binrc`.", name)
		}

		name = p
	}

	nwo := strings.SplitN(name, "/", 2)
	if versionString == "" {
		ev := os.Getenv(fmt.Sprintf("%s_VERSION", strings.ToUpper(nwo[1])))
		if ev == "" {
			return nil, errors.Errorf("unknown project version for %s", name)
		}
		versionString = ev
	}

	if !strings.HasPrefix(versionString, "v") {
		versionString = "v" + versionString
	}

	cleanVersionString := strings.TrimLeft(versionString, "v")
	cleanVersion, err := version.NewVersion(cleanVersionString)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid version %s", versionString)
	}

	var t *template
	projectTemplates, ok := c.templates[nwo[1]]
	if ok {
		for _, tmpl := range projectTemplates {
			constraint, err := version.NewConstraint(tmpl.Range)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid constraint %s", tmpl.Range)
			}
			if constraint.Check(cleanVersion) {
				t = &tmpl
				break
			}
		}

		if t == nil {
			return nil, errors.Errorf("%s's version %s doesn't match any known constraint, binrc cannot install it", nwo[1], versionString)
		}
	} else {
		t = defaultTemplate
	}

	return &Project{
		FullName:     name,
		Version:      versionString,
		Owner:        nwo[0],
		Name:         nwo[1],
		cleanVersion: cleanVersionString,
		template:     t,
	}, nil
}

// New initializes the cache with
// a given store path.
func New(storePath string) (*Cache, error) {
	t, err := loadVersionTemplates()
	if err != nil {
		return nil, err
	}
	return &Cache{
		templates: *t,
		storePath: storePath,
	}, nil
}

// GetOrSet fetches a project from the cache.
// If the project is not in the cache, it tries
// to donwload it from GitHub releases.
func (c *Cache) GetOrSet(name, version string) (*Project, error) {
	project, err := c.newProject(name, version)
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

func loadVersionTemplates() (*templates, error) {
	var r io.Reader
	if fromEnv := os.Getenv("BINRC_TEMPLATES"); fromEnv != "" {
		f, err := os.Open(fromEnv)
		if err != nil {
			return nil, err
		}
		r = f
	} else {
		statikFS, err := fs.New()
		f, err := statikFS.Open("/templates.toml")
		if err != nil {
			return nil, errors.Wrapf(err, "unable to load templates files")
		}
		r = f
	}

	t := &templates{}
	_, err := toml.DecodeReader(r, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
