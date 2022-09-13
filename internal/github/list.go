package github

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gocolly/colly"
)

func ListOfficial() {
	c := colly.NewCollector()
	templates := []string{}
	c.OnHTML("div[aria-labelledby=\"files\"]", func(h *colly.HTMLElement) {
		h.ForEach("div[role=\"row\"]", func(_ int, e *colly.HTMLElement) {
			if e.ChildAttr("div[role=\"gridcell\"] > svg", "aria-label") == "Directory" {
				templates = append(templates, e.ChildAttr("div[role=\"rowheader\"] > span > a", "title"))
			}
		})
	})

	c.Visit("https://github.com/goboiler/templates")

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 11, '\t', tabwriter.AlignRight)
	filled := make([]string, len(templates)+3-len(templates)%3)
	copy(filled, templates)
	for i := 0; i < len(filled); i += 3 {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", filled[i], filled[i+1], filled[i+2])
	}
	writer.Flush()
}
