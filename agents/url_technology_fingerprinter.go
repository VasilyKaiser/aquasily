package agents

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"

	"github.com/VasilyKaiser/aquasily/core"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

// URLTechnologyFingerprinter structure
type URLTechnologyFingerprinter struct {
	session      *core.Session
	technologies map[string]struct{}
}

// ID returns name of the source file
func (a *URLTechnologyFingerprinter) ID() string {
	return "agent:url_technology_fingerprinter"
}

// NewURLTechnologyFingerprinter returns URLTechnologyFingerprinter structure
func NewURLTechnologyFingerprinter() *URLTechnologyFingerprinter {
	return &URLTechnologyFingerprinter{}
}

// Register is registering for EventBus URLResponsive events
func (a *URLTechnologyFingerprinter) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.URLResponsive, a.OnURLResponsive, false)
	a.session = s

	return nil
}

// OnURLResponsive makes request and takes fingerprints
func (a *URLTechnologyFingerprinter) OnURLResponsive(url string) {
	a.session.Out.Debug("[%s] Received new responsive URL %s\n", a.ID(), url)
	page := a.session.GetPage(url)
	if page == nil {
		a.session.Out.Error("[%s] Unable to find page for URL: %s\n", a.ID(), url)
		return
	}
	a.session.WaitGroup.Add()
	go func(page *core.Page) {
		defer a.session.WaitGroup.Done()
		seen := make(map[string]struct{})
		a.fingerprint(page)
		for key := range a.technologies {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			page.AddTag(key, "info", a.constructURL(key))
		}
	}(page)
}

func (a *URLTechnologyFingerprinter) fingerprint(page *core.Page) {
	var body []byte
	var headers http.Header
	var err error
	if page.BodyPath != "" {
		body, err = a.session.ReadFile(page.BodyPath)
		if err != nil {
			a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
			return
		}
		headers = a.getHeaders(page.HeadersPath)
		if headers == nil {
			a.session.Out.Warn("[%s]: Couldn't get headers from the file: %s\n", a.ID(), page.HeadersPath)
			return
		}
	} else {
		// Ignore Certificate Errors
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err := http.DefaultClient.Get(page.URL)
		if err != nil {
			a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
			return
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
			return
		}
		defer resp.Body.Close()
		headers = resp.Header
		page.PageTitle = ExtractTitle(body)
	}

	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
		return
	}
	a.technologies = wappalyzerClient.Fingerprint(headers, body)
	a.session.Out.Debug("[%s] Identified technology %s on %s\n", a.ID(), a.technologies, page.URL)
}

func (a *URLTechnologyFingerprinter) getHeaders(headersPath string) (headers http.Header) {
	headers = make(map[string][]string)
	data, err := a.session.ReadFile(headersPath)
	if err != nil {
		return
	}
	var fileContent []string
	if strings.Contains(string(data), "\r\n") {
		fileContent = strings.Split(string(data), "\r\n")
	} else {
		fileContent = strings.Split(string(data), "\n")
	}
	for i, line := range fileContent {
		if i > 0 && strings.Trim(line, " ") != "" {
			header := strings.Split(line, ": ")
			for j := 0; j < len(header)-1; j++ {
				headers.Add(header[0], header[j+1])
			}
		}
	}
	return
}

func (a *URLTechnologyFingerprinter) constructURL(technologyName string) string {
	wappalyzerURL := "https://www.wappalyzer.com/technologies/"
	techPath, ok := a.session.Technologies[technologyName]
	if ok {
		return wappalyzerURL + techPath[0] + "/" + techPath[1]
	}
	return wappalyzerURL
}
