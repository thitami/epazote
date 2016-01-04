package epazote

import (
	"fmt"
	"github.com/nbari/epazote/scheduler"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const herb = "\U0001f33f"

type Epazote struct {
	Config   Config
	Services Services
}

type Config struct {
	SMTP Email `yaml:"smtp"`
	HTTP Http  `yaml:"http"`
	Scan Scan  `yaml:"scan"`
}

type Email struct {
	Username string
	Password string
	Host     string
	Port     int
	Tls      bool
	Headers  map[string]string
}

type Http struct {
	Host     string
	pathPort int
}

type Every struct {
	Seconds, Minutes, Hours int
}

type Scan struct {
	Paths []string
	Every `yaml:",inline"`
}

type Services map[string]Service

type Service struct {
	URL      string
	Timeout  int
	Every    `yaml:",inline"`
	Log      string
	Expect   Expect
	IfStatus map[int]Action    `yaml:"if_status"`
	IfHeader map[string]Action `yaml:"if_header"`
}

type Expect struct {
	Status int
	Header map[string]string
	Body   interface{}
	IfNot  Action `yaml:"if_not"`
}

type Action struct {
	Cmd    string
	Notify string
	Msg    string
}

func New(file string) (*Epazote, error) {
	yml_file, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var ez Epazote

	if err := yaml.Unmarshal(yml_file, &ez); err != nil {
		return nil, err
	}

	return &ez, nil
}

// CheckPaths verify that directories exist and are readable
func (self *Epazote) CheckPaths() error {
	if len(self.Config.Scan.Paths) > 0 {
		for k, d := range self.Config.Scan.Paths {
			if _, err := os.Stat(d); os.IsNotExist(err) {
				return fmt.Errorf("Verify that directory: %s, exists and is readable.", d)
			}
			if r, err := filepath.EvalSymlinks(d); err != nil {
				return err
			} else {
				self.Config.Scan.Paths[k] = r
			}
		}
		return nil
	}
	return nil
}

// VerifyUrls, we can't supervice unreachable services
func (self *Epazote) VerifyUrls() error {
	ch := AsyncGet(self.Services)
	for i := 0; i < len(self.Services); i++ {
		x := <-ch
		if x.Err != nil {
			return fmt.Errorf("%s - Verify URL: %q", Red(x.Service), x.Err)
		}
	}
	return nil
}

// PathOrServices check if at least one path or service is set
func (self *Epazote) PathsOrServices() error {
	if len(self.Config.Scan.Paths) == 0 && len(self.Services) == 0 {
		return fmt.Errorf("%s", Red("No services to supervices or paths to scan."))
	}
	return nil
}

// Start Add services to scheduler
func (self *Epazote) Start(sk *scheduler.Scheduler) {
	for k, v := range self.Services {
		// Status
		if v.Expect.Status < 1 {
			v.Expect.Status = 200
		}

		// rxBody
		if body, ok := v.Expect.Body.(string); ok {
			re := regexp.MustCompile(body)
			v.Expect.Body = *re
		}

		// schedule service
		sk.AddScheduler(k, GetInterval(60, v.Every), Supervice(v))
	}

	if len(self.Config.Scan.Paths) > 0 {
		s := new(Scandir)
		for _, v := range self.Config.Scan.Paths {
			sk.AddScheduler(v, GetInterval(300, self.Config.Scan.Every), s.Scan(v))
		}
	}

	log.Printf("Epazote %s   on %d services, scan paths: %s [pid: %d]", herb, len(self.Services), strings.Join(self.Config.Scan.Paths, ","), os.Getpid())

	// stop until signal received
	start := time.Now()

	// loop forever
	block := make(chan os.Signal)

	signal.Notify(block, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		signalType := <-block
		switch signalType {
		case syscall.SIGUSR1, syscall.SIGUSR2:
			y, err := yaml.Marshal(&self)
			if err != nil {
				log.Printf("error: %v", err)
			}
			l := `
    Gorutines: %d"
    Alloc : %d
    Total Alloc: %d
    Sys: %d
    Lookups: %d
    Mallocs: %d
    Frees: %d
    Started on: %v
    Uptime: %v`

			runtime.NumGoroutine()
			s := new(runtime.MemStats)
			runtime.ReadMemStats(s)

			log.Printf("Config dump:\n%s---"+l, y, runtime.NumGoroutine(), s.Alloc, s.TotalAlloc, s.Sys, s.Lookups, s.Mallocs, s.Frees, start.Format(time.RFC3339), time.Since(start))

		default:
			signal.Stop(block)
			log.Printf("%q signal received.", signalType)
			sk.StopAll()
			log.Println("Exiting.")
			os.Exit(0)
		}
	}
}

// GetInterval return the check interval in seconds
func GetInterval(d int, e Every) int {
	// default to 60 seconds
	if d < 1 {
		d = 60
	}

	every := d

	if e.Seconds > 0 {
		return e.Seconds
	} else if e.Minutes > 0 {
		return 60 * e.Minutes
	} else if e.Hours > 0 {
		return 3600 * e.Hours
	}

	return every
}

func ParseScan(file string) (Services, error) {
	yml_file, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var s Services

	if err := yaml.Unmarshal(yml_file, &s); err != nil {
		return nil, err
	}

	if len(s) == 0 {
		return nil, fmt.Errorf("[%s] No services found.", Red(file))
	}

	return s, nil
}