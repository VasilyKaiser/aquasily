package agents

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	ctx, cancelExec := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelExec()
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, takeScreenshot(page.URL, &buf)); err != nil {
		a.session.Out.Debug("[%s] Error: %v\n", a.ID(), err)
		a.session.Stats.IncrementScreenshotFailed()
		a.session.Out.Error("%s: screenshot failed: %s\n", page.URL, err)
		return
	}
	if err := os.WriteFile(a.session.GetFilePath(filePath), buf, 0o644); err != nil {
		a.session.Out.Debug("[%s] Error: %v\n", a.ID(), err)
		a.session.Stats.IncrementScreenshotFailed()
		a.session.Out.Error("%s: screenshot failed: %s\n", page.URL, err)
		return
	}
	a.session.Stats.IncrementScreenshotSuccessful()
	a.session.Out.Info("%s: %s\n", page.URL, Green("screenshot successful"))
	page.ScreenshotPath = filePath
	page.HasScreenshot = true
}

func takeScreenshot(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.CaptureScreenshot(res),
	}
}
