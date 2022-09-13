package github

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	ppath "path"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/schollz/progressbar/v3"
)

type irepo interface {
	FetchContent() []content
	fetch() []content
	Download(string)
}

type content_type string

const (
	dir  content_type = "dir"
	file content_type = "file"
)

type content struct {
	_type content_type
	name  string
	url   string
}

type repo struct {
	url     string
	content []content
}

func (r *repo) fetch() []content {
	c := colly.NewCollector()

	c.OnHTML("div[aria-labelledby=\"files\"]", func(h *colly.HTMLElement) {
		wg := &sync.WaitGroup{}
		h.ForEach("div[role=\"row\"]", func(_ int, e *colly.HTMLElement) {
			wg.Add(1)
			go func(e *colly.HTMLElement) {
				if e.ChildAttr("div[role=\"gridcell\"] > svg", "aria-label") == "Directory" {
					url := "https://github.com" + e.ChildAttr("div[role=\"rowheader\"] > span > a", "href")
					newRepo := NewRepo(url)
					_c := newRepo.fetch()
					for i := 0; i < len(_c); i++ {
						_c[i].name = e.ChildAttr("div[role=\"rowheader\"] > span > a", "title") + "/" + _c[i].name
					}

					r.content = append(r.content, _c...)
				} else if e.ChildAttr("div[role=\"gridcell\"] > svg", "aria-label") == "File" {
					title := e.ChildAttr("div[role=\"rowheader\"] > span > a", "title")
					url := "https://github.com" + e.ChildAttr("div[role=\"rowheader\"] > span > a", "href")
					url = strings.Replace(url, "/blob/", "/raw/", 1)
					r.content = append(r.content, content{file, title, url})
				}
				wg.Done()
			}(e)
		})

		wg.Wait()
	})

	c.Visit(r.url)

	return r.content
}

func (r *repo) FetchContent() []content {
	fmt.Println("Fetching files from", r.url)
	return r.fetch()
}

func (r *repo) Download(path string) {
	fmt.Println("Downloading template", r.url)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	os.Chdir(path)

	wg := &sync.WaitGroup{}

	bar := progressbar.Default(int64(len(r.content)))
	for _, c := range r.content {
		wg.Add(1)
		go func(c content) {
			if _, err := os.Stat(ppath.Dir(c.name)); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(ppath.Dir(c.name), os.ModePerm)
				if err != nil {
					panic(err)
				}
			}

			f, err := os.Create(c.name)
			if err != nil {
				panic(err)
			}

			client := &http.Client{}
			req, err := http.NewRequest("GET", c.url, nil)
			if err != nil {
				panic(err)
			}

			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			defer res.Body.Close()
			content, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			_, err = f.Write(content)
			if err != nil {
				panic(err)
			}
			bar.Add(1)
			wg.Done()
		}(c)
	}

	wg.Wait()
	bar.Finish()
	fmt.Println("Downloaded template", r.url)
}

func NewRepo(url string) irepo {
	return &repo{url, []content{}}
}

func NewWorker(repo irepo, download_path string) {
	repo.FetchContent()
	repo.Download(download_path)
}
