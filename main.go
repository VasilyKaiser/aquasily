package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/VasilyKaiser/aquasily/agents"
	"github.com/VasilyKaiser/aquasily/core"
	"github.com/VasilyKaiser/aquasily/parsers"
	"github.com/google/uuid"
)

var (
	sess        *core.Session
	err         error
	reportHTML  = "aquasily_report.html"
	urlsTXT     = "aquasily_urls.txt"
	sessionJSON = "aquasily_session.json"
)

func isURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	if u.Scheme == "" {
		return false
	}
	return true
}

func hasSupportedScheme(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return true
	}
	return false
}

func main() {
	if sess, err = core.NewSession(); err != nil {
		sess.Out.Fatal("Couldn't start a new session: %s", err.Error())
	}

	if *sess.Options.Version {
		sess.Out.Info("%s v%s", core.Name, core.Version)
		os.Exit(0)
	}

	fi, err := os.Stat(*sess.Options.OutDir)

	if os.IsNotExist(err) {
		sess.Out.Fatal("Output destination %s does not exist\n", *sess.Options.OutDir)
	}

	if !fi.IsDir() {
		sess.Out.Fatal("Output destination must be a directory\n")
	}

	sess.Out.Important("%s v%s started at %s\n\n", core.Name, core.Version, sess.Stats.StartedAt.Format(time.RFC3339))

	if *sess.Options.SessionPath != "" {
		jsonSession, err := ioutil.ReadFile(*sess.Options.SessionPath)
		if err != nil {
			sess.Out.Fatal("Unable to read session file at %s: %s\n", *sess.Options.SessionPath, err)
		}

		var parsedSession core.Session
		if err := json.Unmarshal(jsonSession, &parsedSession); err != nil {
			sess.Out.Fatal("Unable to parse session file at %s: %s\n", *sess.Options.SessionPath, err)
		}

		sess.Out.Important("Loaded Aquasily session at %s\n", *sess.Options.SessionPath)
		sess.Out.Important("Generating HTML report...")
		var template []byte
		// If template not specified, use one embeded in bindata.go
		if *sess.Options.TemplatePath != "" {
			template, err = ioutil.ReadFile(*sess.Options.TemplatePath)
		} else {
			template, err = sess.Asset("static/report_template.html")
		}

		if err != nil {
			sess.Out.Fatal("Can't read report template file\n")
		}

		report := core.NewReport(&parsedSession, string(template))
		f, err := os.OpenFile(sess.GetFilePath(reportHTML), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			sess.Out.Fatal("Error during report generation: %s\n", err)
		}

		err = report.Render(f)
		if err != nil {
			sess.Out.Fatal("Error during report generation: %s\n", err)
		}
		sess.Out.Important(" done\n\n")
		sess.Out.Important("Wrote HTML report to: %s\n\n", sess.GetFilePath(reportHTML))
		os.Exit(0)
	}

	agents.NewTCPPortScanner().Register(sess)
	agents.NewURLPublisher().Register(sess)
	agents.NewURLRequester().Register(sess)
	agents.NewURLHostnameResolver().Register(sess)
	agents.NewURLPageTitleExtractor().Register(sess)
	agents.NewURLScreenshotter().Register(sess)
	agents.NewURLTechnologyFingerprinter().Register(sess)

	var targets []string
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		sess.Out.Debug("Data is piped to stdin\n")
		// Reading piped input
		reader := bufio.NewReader(os.Stdin)
		if *sess.Options.Nmap {
			parser := parsers.NewNmapParser()
			targets, err = parser.Parse(reader)
			if err != nil {
				sess.Out.Fatal("Unable to parse input as Nmap/Masscan XML: %s\n", err)
			}
		} else {
			parser := parsers.NewRegexParser()
			targets, err = parser.Parse(reader)
			if err != nil {
				sess.Out.Fatal("Unable to parse input.\n")
			}
		}
	} else {
		sess.Out.Fatal("Feed me with hosts/urls using pipe!\n")
	}

	if len(targets) == 0 {
		sess.Out.Fatal("No targets found in input.\n")
	}

	sess.Out.Important("Targets    : %d\n", len(targets))
	sess.Out.Important("Threads    : %d\n", *sess.Options.Threads)
	sess.Out.Important("Ports      : %s\n", strings.Trim(strings.Replace(fmt.Sprint(sess.Ports), " ", ", ", -1), "[]"))
	sess.Out.Important("Output dir : %s\n\n", *sess.Options.OutDir)

	sess.EventBus.Publish(core.SessionStart)

	for _, target := range targets {
		if isURL(target) {
			if hasSupportedScheme(target) {
				sess.EventBus.Publish(core.URL, target)
			}
		} else {
			sess.EventBus.Publish(core.Host, target)
		}
	}

	time.Sleep(1 * time.Second)
	sess.EventBus.WaitAsync()
	sess.WaitGroup.Wait()

	sess.EventBus.Publish(core.SessionEnd)
	time.Sleep(2 * time.Second)
	sess.EventBus.WaitAsync()
	sess.WaitGroup.Wait()

	sess.Out.Important("Calculating page structures...")
	f, _ := os.OpenFile(sess.GetFilePath(urlsTXT), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	for _, page := range sess.Pages {
		filename := sess.GetFilePath(fmt.Sprintf("html/%s.html", page.BaseFilename()))
		body, err := os.Open(filename)
		if err != nil {
			continue
		}
		structure, _ := core.GetPageStructure(body)
		page.PageStructure = structure
		f.WriteString(page.URL + "\n")
	}
	f.Close()
	sess.Out.Important(" done\n")

	sess.Out.Important("Clustering similar pages...")
	for _, page := range sess.Pages {
		foundCluster := false
		for clusterUUID, cluster := range sess.PageSimilarityClusters {
			addToCluster := true
			for _, pageURL := range cluster {
				page2 := sess.GetPage(pageURL)
				if page2 != nil && core.GetSimilarity(page.PageStructure, page2.PageStructure) < 0.80 {
					addToCluster = false
					break
				}
			}

			if addToCluster {
				foundCluster = true
				sess.PageSimilarityClusters[clusterUUID] = append(sess.PageSimilarityClusters[clusterUUID], page.URL)
				break
			}
		}

		if !foundCluster {
			newClusterUUID := uuid.New().String()
			sess.PageSimilarityClusters[newClusterUUID] = []string{page.URL}
		}
	}
	sess.Out.Important(" done\n")

	sess.Out.Important("Generating HTML report...")
	var template []byte
	if *sess.Options.TemplatePath != "" {
		template, err = ioutil.ReadFile(*sess.Options.TemplatePath)
	} else {
		template, err = sess.Asset("static/report_template.html")
	}

	if err != nil {
		sess.Out.Fatal("Can't read report template file\n")
	}
	report := core.NewReport(sess, string(template))
	f, err = os.OpenFile(sess.GetFilePath(reportHTML), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		sess.Out.Fatal("Error during report generation: %s\n", err)
	}
	err = report.Render(f)
	if err != nil {
		sess.Out.Fatal("Error during report generation: %s\n", err)
	}
	sess.Out.Important(" done\n\n")

	sess.End()

	sess.Out.Important("Writing session file...")
	err = sess.SaveToFile(sessionJSON)
	if err != nil {
		sess.Out.Error("Failed!\n")
		sess.Out.Debug("Error: %v\n", err)
	}

	sess.Out.Important("Time:\n")
	sess.Out.Info(" - Started at  : %v\n", sess.Stats.StartedAt.Format(time.RFC3339))
	sess.Out.Info(" - Finished at : %v\n", sess.Stats.FinishedAt.Format(time.RFC3339))
	sess.Out.Info(" - Duration    : %v\n\n", sess.Stats.Duration().Round(time.Second))

	sess.Out.Important("Requests:\n")
	sess.Out.Info(" - Successful : %v\n", sess.Stats.RequestSuccessful)
	sess.Out.Info(" - Failed     : %v\n\n", sess.Stats.RequestFailed)

	sess.Out.Info(" - 2xx : %v\n", sess.Stats.ResponseCode2xx)
	sess.Out.Info(" - 3xx : %v\n", sess.Stats.ResponseCode3xx)
	sess.Out.Info(" - 4xx : %v\n", sess.Stats.ResponseCode4xx)
	sess.Out.Info(" - 5xx : %v\n\n", sess.Stats.ResponseCode5xx)

	sess.Out.Important("Screenshots:\n")
	sess.Out.Info(" - Successful : %v\n", sess.Stats.ScreenshotSuccessful)
	sess.Out.Info(" - Failed     : %v\n\n", sess.Stats.ScreenshotFailed)

	sess.Out.Important("Wrote HTML report to: %s\n\n", sess.GetFilePath(reportHTML))
}
