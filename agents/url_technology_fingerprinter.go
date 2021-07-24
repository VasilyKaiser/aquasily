package agents

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/VasilyKaiser/aquasily/core"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

type URLTechnologyFingerprinter struct {
	session      *core.Session
	technologies map[string]struct{}
}

func (a *URLTechnologyFingerprinter) ID() string {
	return "agent:url_technology_fingerprinter"
}

func NewURLTechnologyFingerprinter() *URLTechnologyFingerprinter {
	return &URLTechnologyFingerprinter{}
}

func (a *URLTechnologyFingerprinter) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.URLResponsive, a.OnURLResponsive, false)
	a.session = s

	return nil
}

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
			page.AddTag(key, "info", page.URL)
		}
	}(page)
}

func (a *URLTechnologyFingerprinter) fingerprint(page *core.Page) {
	// Ignore Certificate Errors
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.DefaultClient.Get(page.URL)
	if err != nil {
		a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
		return
	}
	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		a.session.Out.Error("[%s]: %s\n", a.ID(), err.Error())
		return
	}
	a.technologies = wappalyzerClient.Fingerprint(resp.Header, data)
	a.session.Out.Debug("[%s] Identified technology %s on %s\n", a.ID(), a.technologies, page.URL)
}
