package core

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/asaskevich/EventBus"
)

// Variables for Session handling
const (
	SessionStart  = "session:start"
	SessionEnd    = "session:end"
	Host          = "host"
	URL           = "url"
	URLResponsive = "url:responsive"
	TCPPort       = "port:tcp"
)

// Stats structure
type Stats struct {
	StartedAt            time.Time `json:"startedAt"`
	FinishedAt           time.Time `json:"finishedAt"`
	PortOpen             uint32    `json:"portOpen"`
	PortClosed           uint32    `json:"portClosed"`
	RequestSuccessful    uint32    `json:"requestSuccessful"`
	RequestFailed        uint32    `json:"requestFailed"`
	ResponseCode2xx      uint32    `json:"responseCode2xx"`
	ResponseCode3xx      uint32    `json:"responseCode3xx"`
	ResponseCode4xx      uint32    `json:"responseCode4xx"`
	ResponseCode5xx      uint32    `json:"responseCode5xx"`
	ScreenshotSuccessful uint32    `json:"screenshotSuccessful"`
	ScreenshotFailed     uint32    `json:"screenshotFailed"`
}

// Duration returns duration
func (s *Stats) Duration() time.Duration {
	return s.FinishedAt.Sub(s.StartedAt)
}

// IncrementPortOpen increments number of ports open
func (s *Stats) IncrementPortOpen() {
	atomic.AddUint32(&s.PortOpen, 1)
}

// IncrementPortClosed increments number of ports closed
func (s *Stats) IncrementPortClosed() {
	atomic.AddUint32(&s.PortClosed, 1)
}

// IncrementRequestSuccessful increments number of successful requests
func (s *Stats) IncrementRequestSuccessful() {
	atomic.AddUint32(&s.RequestSuccessful, 1)
}

// IncrementRequestFailed increments number of requests failed
func (s *Stats) IncrementRequestFailed() {
	atomic.AddUint32(&s.RequestFailed, 1)
}

// IncrementResponseCode2xx ...
func (s *Stats) IncrementResponseCode2xx() {
	atomic.AddUint32(&s.ResponseCode2xx, 1)
}

// IncrementResponseCode3xx ...
func (s *Stats) IncrementResponseCode3xx() {
	atomic.AddUint32(&s.ResponseCode3xx, 1)
}

// IncrementResponseCode4xx ...
func (s *Stats) IncrementResponseCode4xx() {
	atomic.AddUint32(&s.ResponseCode4xx, 1)
}

// IncrementResponseCode5xx ...
func (s *Stats) IncrementResponseCode5xx() {
	atomic.AddUint32(&s.ResponseCode5xx, 1)
}

// IncrementScreenshotSuccessful ...
func (s *Stats) IncrementScreenshotSuccessful() {
	atomic.AddUint32(&s.ScreenshotSuccessful, 1)
}

// IncrementScreenshotFailed ...
func (s *Stats) IncrementScreenshotFailed() {
	atomic.AddUint32(&s.ScreenshotFailed, 1)
}

// Session structure
type Session struct {
	sync.Mutex
	Version                string              `json:"version"`
	Options                Options             `json:"-"`
	Out                    *Logger             `json:"-"`
	Stats                  *Stats              `json:"stats"`
	Pages                  map[string]*Page    `json:"pages"`
	PageSimilarityClusters map[string][]string `json:"pageSimilarityClusters"`
	Ports                  []int               `json:"-"`
	EventBus               EventBus.Bus        `json:"-"`
	WaitGroup              SizedWaitGroup      `json:"-"`
}

// Start initiating all required tasks
func (s *Session) Start() {
	s.Pages = make(map[string]*Page)
	s.PageSimilarityClusters = make(map[string][]string)
	s.initStats()
	s.initLogger()
	s.initPorts()
	s.initThreads()
	s.initEventBus()
	s.initWaitGroup()
}

// End reports time finished
func (s *Session) End() {
	s.Stats.FinishedAt = time.Now()
}

// AddPage returns page and nil or nil and err if error occure
func (s *Session) AddPage(url string) (*Page, error) {
	s.Lock()
	defer s.Unlock()
	if page, ok := s.Pages[url]; ok {
		return page, nil
	}

	page, err := NewPage(url)
	if err != nil {
		return nil, err
	}

	s.Pages[url] = page
	return page, nil
}

// GetPage returns page from Session Pages map if exists or nil
func (s *Session) GetPage(url string) *Page {
	if page, ok := s.Pages[url]; ok {
		return page
	}
	return nil
}

// GetPageByUUID returns page matching by UUID or nil
func (s *Session) GetPageByUUID(id string) *Page {
	for _, page := range s.Pages {
		if page.UUID == id {
			return page
		}
	}
	return nil
}

func (s *Session) initStats() {
	if s.Stats != nil {
		return
	}
	s.Stats = &Stats{
		StartedAt: time.Now(),
	}
}

func (s *Session) initPorts() {
	var ports []int
	switch *s.Options.Ports {
	case "small":
		ports = SmallPortList
	case "", "medium", "default":
		ports = MediumPortList
	case "large":
		ports = LargePortList
	case "xlarge", "huge":
		ports = XLargePortList
	default:
		for _, p := range strings.Split(*s.Options.Ports, ",") {
			port, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				s.Out.Fatal("Invalid port range given: %s\n", err.Error())
			}
			if port < 1 || port > 65535 {
				s.Out.Fatal("Invalid port given: %v: %s\n", port, err.Error())
			}
			ports = append(ports, port)
		}
	}
	s.Ports = ports
}

func (s *Session) initLogger() {
	s.Out = &Logger{}
	s.Out.SetDebug(*s.Options.Debug)
	s.Out.SetSilent(*s.Options.Silent)
}

func (s *Session) initThreads() {
	if *s.Options.Threads == 0 {
		numCPUs := runtime.NumCPU()
		s.Options.Threads = &numCPUs
	}
}

func (s *Session) initEventBus() {
	s.EventBus = EventBus.New()
}

func (s *Session) initWaitGroup() {
	s.WaitGroup = New(*s.Options.Threads)
}

// InitDirectories makes needed directories inside OutPath
func (s *Session) InitDirectories() {
	*s.Options.OutDir = path.Join(*s.Options.OutDir, getFolderName())
	if !checkIfPathExists(*s.Options.OutDir) {
		for i := 1; i < 10000; i++ {
			newFolderName := *s.Options.OutDir + "(" + fmt.Sprint(i) + ")"
			if checkIfPathExists(newFolderName) {
				*s.Options.OutDir = newFolderName
				break
			}
		}
	}
	for _, d := range []string{"headers", "html", "screenshots"} {
		d = s.GetFilePath(d)
		if _, err := os.Stat(d); os.IsNotExist(err) {
			err = os.MkdirAll(d, 0755)
			if err != nil {
				s.Out.Fatal("Failed to create required directory %s: %s\n", d, err.Error())
			}
		}
	}
}

func checkIfPathExists(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func getFolderName() string {
	const layout = "01-02-2006"
	t := time.Now()
	return "aquasilyReport_" + t.Format(layout)
}

// BaseFilenameFromURL returns filename made up from URL
func (s *Session) BaseFilenameFromURL(stru string) string {
	u, err := url.Parse(stru)
	if err != nil {
		return ""
	}

	h := sha1.New()
	io.WriteString(h, u.Path)
	io.WriteString(h, u.Fragment)

	pathHash := fmt.Sprintf("%x", h.Sum(nil))[0:16]
	host := strings.Replace(u.Host, ":", "__", 1)
	filename := fmt.Sprintf("%s__%s__%s", u.Scheme, strings.Replace(host, ".", "_", -1), pathHash)
	return strings.ToLower(filename)
}

// GetFilePath returns path joined with OutDir
func (s *Session) GetFilePath(p string) string {
	return path.Join(*s.Options.OutDir, p)
}

// ReadFile returns content of the file
func (s *Session) ReadFile(p string) ([]byte, error) {
	content, err := os.ReadFile(s.GetFilePath(p))
	if err != nil {
		return content, err
	}
	return content, nil
}

// ToJSON returns session in JSON format
func (s *Session) ToJSON() string {
	sessionJSON, _ := json.Marshal(s)
	return string(sessionJSON)
}

// SaveToFile saves Session to the file
func (s *Session) SaveToFile(filename string) error {
	path := s.GetFilePath(filename)
	err := os.WriteFile(path, []byte(s.ToJSON()), 0644)
	if err != nil {
		return err
	}

	return nil
}

// NewSession sets options and returns a new session
func NewSession() (*Session, error) {
	var err error
	var session Session

	session.Version = Version

	if session.Options, err = ParseOptions(); err != nil {
		return nil, err
	}

	dest, err := os.Stat(*session.Options.OutDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("output destination %s does not exist", *session.Options.OutDir)
	}
	if !dest.IsDir() {
		return nil, fmt.Errorf("output destination must be a directory: %s", *session.Options.OutDir)
	}

	if *session.Options.BrowserPath != "" {
		if _, err := os.Stat(*session.Options.BrowserPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("chrome path %s does not exist", *session.Options.BrowserPath)
		}
	}

	if *session.Options.SessionPath != "" {
		if _, err := os.Stat(*session.Options.SessionPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("session path %s does not exist", *session.Options.SessionPath)
		}
	}

	if *session.Options.TemplatePath != "" {
		if _, err := os.Stat(*session.Options.TemplatePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("template path %s does not exist", *session.Options.TemplatePath)
		}
	}

	envOutPath := os.Getenv("AQUASILY_OUT_PATH")
	if *session.Options.OutDir == "." && envOutPath != "" {
		session.Options.OutDir = &envOutPath
	}

	outdir := filepath.Clean(*session.Options.OutDir)
	session.Options.OutDir = &outdir

	session.Version = Version
	session.Start()

	return &session, nil
}
