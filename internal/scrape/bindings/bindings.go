// Package bindings exposes the scraper to Python via gopy (make gopy).
package bindings

import (
	"github.com/GarroshIcecream/yummy/internal/scrape"
)

// ScrapeURLToJSON returns recipe JSON for url. pythonPath optional (e.g. "" for default). On error err is non-empty.
func ScrapeURLToJSON(url string, pythonPath string) (json string, err string) {
	s, e := scrape.ScrapeURLRaw(url, pythonPath)
	if e != nil {
		return "", e.Error()
	}
	return s, ""
}
