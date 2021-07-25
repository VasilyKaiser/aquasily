package agents

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/VasilyKaiser/aquasily/core"
)

// URLPublisher structure
type URLPublisher struct {
	session *core.Session
}

// NewURLPublisher returns URLPublisher structure
func NewURLPublisher() *URLPublisher {
	return &URLPublisher{}
}

// ID returns name of the source file
func (a *URLPublisher) ID() string {
	return "agent:url_publisher"
}

// Register is registering for EventBus TCPPort events
func (a *URLPublisher) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.TCPPort, a.OnTCPPort, false)
	a.session = s
	return nil
}

// OnTCPPort constructs URL and publishes it to EventBus
func (a *URLPublisher) OnTCPPort(port int, host string) {
	a.session.Out.Debug("[%s] Received new open port on %s: %d\n", a.ID(), host, port)
	var url string
	if a.isTLS(port, host) {
		url = HostAndPortToURL(host, port, "https")
	} else {
		url = HostAndPortToURL(host, port, "http")
	}
	a.session.EventBus.Publish(core.URL, url)
}

func (a *URLPublisher) isTLS(port int, host string) bool {
	if port == 80 {
		return false
	}
	if port == 443 {
		return true
	}
	dialer := &net.Dialer{Timeout: time.Duration(*a.session.Options.HTTPTimeout) * time.Millisecond}
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", host, port), conf)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
