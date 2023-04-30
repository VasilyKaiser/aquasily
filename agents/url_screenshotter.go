package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/VasilyKaiser/aquasily/core"
	"github.com/chromedp/chromedp"
)

// URLScreenshotter structure
type URLScreenshotter struct {
	session         *core.Session
	tempUserDirPath string
}

// NewURLScreenshotter returns URLScreenshotter structure
func NewURLScreenshotter() *URLScreenshotter {
	return &URLScreenshotter{}
}

// ID returns name of the source file
func (a *URLScreenshotter) ID() string {
	return "agent:url_screenshotter"
}

// Register is registering for EventBus URLResponsive and SessionEnd events
func (a *URLScreenshotter) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.URLResponsive, a.OnURLResponsive, false)
	s.EventBus.SubscribeAsync(core.SessionEnd, a.OnSessionEnd, false)
	a.session = s
	a.createTempUserDir()

	return nil
}

// OnURLResponsive takes screenshot of the page
func (a *URLScreenshotter) OnURLResponsive(url string) {
	a.session.Out.Debug("[%s] Received new responsive URL %s\n", a.ID(), url)
	page := a.session.GetPage(url)
	if page == nil {
		a.session.Out.Error("Unable to find page for URL: %s\n", url)
		return
	}
	a.session.WaitGroup.Add()
	time.Sleep(2 * time.Second)
	go func(page *core.Page) {
		defer a.session.WaitGroup.Done()
		a.screenshotPage(page)
	}(page)
}

// OnSessionEnd removes temp directory
func (a *URLScreenshotter) OnSessionEnd() {
	a.session.Out.Debug("[%s] Received SessionEnd event\n", a.ID())
	os.RemoveAll(a.tempUserDirPath)
	a.session.Out.Debug("[%s] Deleted temporary user directory at: %s\n", a.ID(), a.tempUserDirPath)
}

func (a *URLScreenshotter) createTempUserDir() {
	dir, err := os.MkdirTemp("", "aquatone-browser")
	if err != nil {
		a.session.Out.Fatal("[%s] Unable to create temporary user directory for Chrome/Chromium browser: %s\n", a.ID(), err.Error())
	}
	a.session.Out.Debug("[%s] Created temporary user directory at: %s\n", a.ID(), dir)
	a.tempUserDirPath = dir
}

func (a *URLScreenshotter) compareScreenshots(page *core.Page) {
	if *a.session.Options.ReferenceScreenshots != "" {
		if val, ok := a.session.ReferenceScreenshots[filepath.Base(page.ScreenshotPath)]; ok {
			page.Different = true
			page.LastScreenshot = val
			absScreenshotPath, _ := filepath.Abs(a.session.GetFilePath(page.ScreenshotPath))
			page.ScreenshotDifference = a.Compare(val, absScreenshotPath)
		} else {
			a.session.Out.Debug("No screenshot to compare!\n")
			page.Different = false
		}
	}
}

func (a *URLScreenshotter) screenshotPage(page *core.Page) {
	filePath := fmt.Sprintf("screenshots/%s.png", page.BaseFilename())
	resolution := strings.Split(*a.session.Options.Resolution, ",")

	width, _ := strconv.Atoi(resolution[0])
	height, _ := strconv.Atoi(resolution[1])

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(RandomUserAgent()),
		chromedp.WindowSize(width, height),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.IgnoreCertErrors,
	}
	if *a.session.Options.BrowserPath != "" {
		opts = append(opts, chromedp.ExecPath(*a.session.Options.BrowserPath))
	}
	// if *a.session.Options.Proxy != "" {
	// 	opts = append(opts, chromedp.ProxyServer(*a.session.Options.Proxy))
	// }
	c1 := make(chan string, 1)
	ctx, cancelExec := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelExec()
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	a.session.Out.Debug("[%v] Attending to capture: %s\n", a.ID(), page.URL)
	var buf []byte
	var outOfTime = false
	var dtstart time.Time
	var dtend time.Time
	go func() {
		dtstart = time.Now()
		capturePart := func() string {
			if err := chromedp.Run(ctx, takeScreenshot(page.URL, &buf)); err != nil {
				if err.Error() != "context canceled" && !outOfTime {
					a.session.Out.Debug("[%s] Error: %v\n", a.ID(), err)
					a.session.Stats.IncrementScreenshotFailed()
					a.session.Out.Error("%s: Screenshot failed: %s\n", page.URL, err)
					cancel()
				}
				return "not done"
			}
			dtend = time.Now()
			if err := os.WriteFile(a.session.GetFilePath(filePath), buf, 0o644); err != nil {
				if !outOfTime {
					a.session.Out.Debug("[%s] Error: %v\n", a.ID(), err)
					a.session.Stats.IncrementScreenshotFailed()
					a.session.Out.Error("%s: Failed to write to file: %s\n", page.URL, err.Error())
					return "not done"
				}
			}
			return "done"
		}()
		c1 <- capturePart
	}()
	select {
	case res := <-c1:
		if res == "done" && outOfTime {
			return
		} else if res == "done" {
			page.ScreenshotPath = filePath
			page.HasScreenshot = true
			a.session.Out.Debug("%s done - screenshot: %v\n", page.URL, page.HasScreenshot)
			a.session.Stats.IncrementScreenshotSuccessful()
			a.session.Out.Info("%s: %s %s\n", page.URL, Green("screenshot successful"), dtend.Sub(dtstart).Round(time.Second))
			cancel()
			a.compareScreenshots(page)
			return
		} else if res == "not done" && outOfTime {
			return
		}
	case <-time.After(time.Duration(*a.session.Options.ScreenshotTimeout) * time.Second):
		outOfTime = true
		if err := chromedp.Run(ctx, capture(&buf)); err != nil {
			a.session.Out.Error("[%s] Error while capturing the screen: %s\n", a.ID(), err.Error())
			a.session.Stats.IncrementScreenshotFailed()
			return
		}
		dtend = time.Now()
		if err := os.WriteFile(a.session.GetFilePath(filePath), buf, 0o644); err != nil {
			a.session.Out.Error("[%s] Error while writing to file: %\n", a.ID(), err.Error())
			a.session.Stats.IncrementScreenshotFailed()
			return
		}
		a.session.Stats.IncrementScreenshotSuccessful()
		a.session.Out.Info("%s: %s %s\n", page.URL, Green("screenshot successful"), dtend.Sub(dtstart).Round(time.Second))
		cancel()
		page.ScreenshotPath = filePath
		page.HasScreenshot = true
		a.session.Out.Debug("%s after timeout - screenshot: %v\n", page.URL, page.HasScreenshot)
		a.compareScreenshots(page)
		return
	}
	page.HasScreenshot = false
	a.session.Out.Debug("%s screenshot: %v\n", page.URL, page.HasScreenshot)
}

func takeScreenshot(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.CaptureScreenshot(res),
	}
}
func capture(res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.CaptureScreenshot(res),
	}
}
