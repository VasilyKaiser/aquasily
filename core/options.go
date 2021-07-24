package core

import (
	"flag"
	"fmt"
	"strings"
)

type Options struct {
	Threads           *int
	OutDir            *string
	SessionPath       *string
	TemplatePath      *string
	Proxy             *string
	BrowserPath       *string
	Resolution        *string
	Ports             *string
	ScanTimeout       *int
	HTTPTimeout       *int
	ScreenshotTimeout *int
	Nmap              *bool
	SaveBody          *bool
	Silent            *bool
	Debug             *bool
	Version           *bool
}

func ParseOptions() (Options, error) {
	options := Options{
		Version:           flag.Bool("version", false, "Print current Aquasily version"),
		OutDir:            flag.String("out", ".", "Directory to write files to"),
		Threads:           flag.Int("threads", 0, "Number of concurrent threads (default number of logical CPUs)"),
		Ports:             flag.String("ports", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(MediumPortList)), ","), "[]"), "Ports to scan on hosts. Supported list aliases: small, medium, large, xlarge"),
		ScanTimeout:       flag.Int("scan-timeout", 600, "Timeout in milliseconds for port scans"),
		Nmap:              flag.Bool("nmap", false, "Parse input as Nmap/Masscan XML"),
		BrowserPath:       flag.String("browser", "", "Full path to the Chrome/Chromium/Edge executable to use. By default, aquasily will search for Chrome or Chromium"),
		Resolution:        flag.String("resolution", "1200,900", "Screenshot resolution"),
		Proxy:             flag.String("proxy", "", "Proxy to use for HTTP requests"),
		HTTPTimeout:       flag.Int("http-timeout", 3*1000, "Timeout in milliseconds for HTTP requests"),
		ScreenshotTimeout: flag.Int("screenshot-timeout", 6*1000, "Timeout in milliseconds for screenshots"),
		Silent:            flag.Bool("silent", false, "Suppress all output except for errors"),
		Debug:             flag.Bool("debug", false, "Print debugging information"),
		SaveBody:          flag.Bool("save-body", true, "Save response bodies to files"),
		SessionPath:       flag.String("session", "", "Load Aquasily session file and generate HTML report"),
		TemplatePath:      flag.String("template", "", "Path to HTML template to use for report"),
	}
	flag.Parse()
	return options, nil
}
