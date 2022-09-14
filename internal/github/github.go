package github

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	ppath "path"
	"strings"
	"sync"

	"github.com/goboiler/gboil-cli/internal/boiler"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
)

type irepo interface {
	FetchContent() *boiler.Gboil
	fetch() *boiler.Gboil
	Download(string)
}

type repo struct {
	url     string
	content *boiler.Gboil
}

func (r *repo) fetch() *boiler.Gboil {
	client := &http.Client{}
	req, err := http.NewRequest("GET", strings.Replace(r.url+"/.gboil.yml", "/tree/", "/raw/", 1), nil)
	if err != nil {
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		return nil
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil
	}

	var y boiler.Gboil
	err = yaml.NewDecoder(res.Body).Decode(&y)
	if err != nil {
		return nil
	}

	r.content = &y
	return &y
}

func (r *repo) FetchContent() *boiler.Gboil {
	fmt.Println("Fetching files from", r.url)
	return r.fetch()
}

func (r *repo) Download(path string) {
	_p := path
	if _p == "" {
		_p = r.content.Name
	}

	if _, err := os.Stat(_p); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(_p, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	os.Chdir(_p)

	fmt.Println("Downloading", r.url, "to", _p)

	wg := &sync.WaitGroup{}
	bar := progressbar.Default(int64(len(r.content.Files)))

	for _, f := range r.content.Files {
		wg.Add(1)
		go func(f *boiler.File) {
			fmt.Println("==> Downloading", f.Path)
			if _, err := os.Stat(ppath.Dir(f.Path)); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(ppath.Dir(f.Path), os.ModePerm)
				if err != nil {
					panic(err)
				}
			}

			fi, err := os.Create(f.Path)
			if err != nil {
				panic(err)
			}
			defer fi.Close()

			client := &http.Client{}
			req, err := http.NewRequest("GET", strings.Replace(r.url+"/"+f.Path, "/tree/", "/raw/", 1), nil)
			if err != nil {
				panic(err)
			}

			res, err := client.Do(req)
			if err != nil || res.StatusCode != 200 {
				panic(err)
			}

			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			_s := sha256.New()
			_s.Write(b)
			_sha256 := fmt.Sprintf("%x", _s.Sum(nil))
			if _sha256 != f.Sha {
				fmt.Println(f.Sha)
				fmt.Println(_sha256)
				panic("sha256 mismatch")
			}

			fi.Write(b)

			bar.Add(1)
			wg.Done()
		}(f)
	}

	wg.Wait()
	bar.Finish()
	fmt.Println("Downloaded template", r.url)
}

func NewRepo(url string) irepo {
	return &repo{url, nil}
}

func NewWorker(repo irepo, download_path string) {
	repo.FetchContent()
	repo.Download(download_path)
}
