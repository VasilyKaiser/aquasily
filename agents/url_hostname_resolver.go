package agents

import (
	"fmt"
	"net"

	"github.com/VasilyKaiser/aquasily/core"
)

// URLHostnameResolver structure
type URLHostnameResolver struct {
	session *core.Session
}

// NewURLHostnameResolver returns URLHostnameResolver structure
func NewURLHostnameResolver() *URLHostnameResolver {
	return &URLHostnameResolver{}
}

// ID returns name of the source file
func (a *URLHostnameResolver) ID() string {
	return "agent:url_hostname_resolver"
}

// Register is registering for EventBus URLResponsive events
func (a *URLHostnameResolver) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.URLResponsive, a.OnURLResponsive, false)
	a.session = s

	return nil
}

// OnURLResponsive is resolving the host/IP provided
func (a *URLHostnameResolver) OnURLResponsive(url string) {
	a.session.Out.Debug("[%s] Received new responsive URL %s\n", a.ID(), url)
	page := a.session.GetPage(url)
	if page == nil {
		a.session.Out.Error("Unable to find page for URL: %s\n", url)
		return
	}
	if page.IsIPHost() {
		a.session.Out.Debug("[%s] Skipping hostname resolving on IP host: %s\n", a.ID(), url)
		page.Addrs = []string{page.ParsedURL().Hostname()}
		return
	}

	a.session.WaitGroup.Add()
	go func(page *core.Page) {
		defer a.session.WaitGroup.Done()
		addrs, err := net.LookupHost(fmt.Sprintf("%s.", page.ParsedURL().Hostname()))
		if err != nil {
			a.session.Out.Debug("[%s] Error: %v\n", a.ID(), err)
			a.session.Out.Error("Failed to resolve hostname for %s\n", page.URL)
			return
		}
		page.Addrs = addrs
	}(page)
}
