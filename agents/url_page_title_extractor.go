package agents

import (
	"bytes"
	"fmt"

	"github.com/VasilyKaiser/aquasily/core"
	"golang.org/x/net/html"
)

type URLPageTitleExtractor struct {
	session *core.Session
}

func NewURLPageTitleExtractor() *URLPageTitleExtractor {
	return &URLPageTitleExtractor{}
}

func (a *URLPageTitleExtractor) ID() string {
	return "agent:url_page_title_extractor"
}

func (a *URLPageTitleExtractor) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.URLResponsive, a.OnURLResponsive, false)
	a.session = s

	return nil
}

func (a *URLPageTitleExtractor) OnURLResponsive(url string) {
	a.session.Out.Debug("[%s] Received new responsive URL %s\n", a.ID(), url)
	page := a.session.GetPage(url)
	if page == nil {
		a.session.Out.Error("Unable to find page for URL: %s\n", url)
		return
	}

	a.session.WaitGroup.Add()
	go func(page *core.Page) {
		defer a.session.WaitGroup.Done()
		body, err := a.session.ReadFile(fmt.Sprintf("html/%s.html", page.BaseFilename()))
		if err != nil {
			a.session.Out.Debug("[%s] Error reading HTML body file for %s: %s\n", a.ID(), page.URL, err)
			return
		}
		doc, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			a.session.Out.Debug("[%s] Error when parsing HTML body file for %s: %s\n", a.ID(), page.URL, err)
			return
		}
		a.session.Out.Debug("Extracting title from: %v ", page.Hostname)
		page.PageTitle = pageTitle(doc)
	}(page)
}

// pageTitle given a reference to a html.Node, scans it until it
// finds the title tag, and returns its value
func pageTitle(n *html.Node) string {
	var title string
	if n.Type == html.ElementNode && n.Data == "title" {
		// Check if title contains something
		if n.FirstChild != nil {
			return n.FirstChild.Data
		} else {
			return ""
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title = pageTitle(c)
		if title != "" {
			break
		}
	}
	return title
}
