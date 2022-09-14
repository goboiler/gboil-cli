package boiler

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type File struct {
	Path string `json:"path" yaml:"path"`
	Sha  string `json:"sha" yaml:"sha"`
}

type Gboil struct {
	Name  string  `json:"name" yaml:"name"`
	Sha   string  `json:"sha" yaml:"sha"`
	Files []*File `json:"files" yaml:"files"`
}

func Gen() {
	abs, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	gb := &Gboil{}
	gb.Name = filepath.Base(abs)

	var tsha []string
	filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(path, ".git/") {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			s := sha256.New()
			s.Write(b)
			sha := fmt.Sprintf("%x", s.Sum(nil))
			tsha = append(tsha, sha)

			gb.Files = append(gb.Files, &File{Path: path, Sha: sha})
		}

		return nil
	})

	ts := sha256.New()
	ts.Write([]byte(strings.Join(tsha, ".")))
	_tsha := fmt.Sprintf("%x", ts.Sum(nil))
	gb.Sha = _tsha

	f, err := os.Create(".gboil.yml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := yaml.NewEncoder(f).Encode(gb); err != nil {
		panic(err)
	}
}
