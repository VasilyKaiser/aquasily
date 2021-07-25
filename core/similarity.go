package core

import (
	"fmt"
	"io"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/net/html"
)

// GetPageStructure returns structure from body provided
func GetPageStructure(body io.Reader) ([]string, error) {
	var structure []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		token := z.Token()
		switch tt {
		case html.ErrorToken:
			return structure, nil
		case html.StartTagToken:
			structure = append(structure, token.Data)
			for _, attr := range token.Attr {
				if attr.Key != "id" {
					continue
				}
				structure = append(structure, fmt.Sprintf("#%s", attr.Val))
				break
			}
		}
	}
}

// GetSimilarity return a measure of the sequences' similarity
func GetSimilarity(a, b []string) float64 {
	matcher := difflib.NewMatcher(a, b)
	return matcher.Ratio()
}
