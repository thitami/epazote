package epazote

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
)

func TestSuperviceStatusCreated(t *testing.T) {
	var wg sync.WaitGroup
	check_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer check_s.Close()
	log_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		decoder := json.NewDecoder(r.Body)
		var i map[string]interface{}
		err := decoder.Decode(&i)
		if err != nil {
			t.Error(err)
		}
		// check name
		if n, ok := i["name"]; ok {
			if n != "s 1" {
				t.Errorf("Expecting  %q, got: %q", "s 1", n)
			}
		} else {
			t.Errorf("key not found: %q", "name")
		}
		// check because
		if b, ok := i["because"]; ok {
			if b != "Status: 201" {
				t.Errorf("Expecting: %q, got: %q", "Status: 201", b)
			}
		} else {
			t.Errorf("key not found: %q", "because")
		}
		// check exit
		if e, ok := i["exit"]; ok {
			if e.(float64) != 0 {
				t.Errorf("Expecting: 0 got: %v", e.(float64))
			}
		} else {
			t.Errorf("key not found: %q", "exit")
		}
		// check url
		if _, ok := i["url"]; !ok {
			t.Error("URL key not found")
		}
		wg.Done()
	}))
	defer log_s.Close()
	s := make(Services)
	s["s 1"] = Service{
		Name: "s 1",
		URL:  check_s.URL,
		Log:  log_s.URL,
		Expect: Expect{
			Status: 201,
		},
	}
	ez := &Epazote{
		Services: s,
	}
	wg.Add(1)
	ez.Supervice(s["s 1"])()
	wg.Wait()
}

func TestSuperviceBodyMatch(t *testing.T) {
	var wg sync.WaitGroup
	check_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		fmt.Fprintln(w, "Hello, epazote match 0BC20225-2E72-4646-9202-8467972199E1 regex")
	}))
	defer check_s.Close()
	log_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		decoder := json.NewDecoder(r.Body)
		var i map[string]interface{}
		err := decoder.Decode(&i)
		if err != nil {
			t.Error(err)
		}
		// check name
		if n, ok := i["name"]; ok {
			if n != "s 1" {
				t.Errorf("Expecting  %q, got: %q", "s 1", n)
			}
		} else {
			t.Errorf("key not found: %q", "name")
		}
		// check because
		if b, ok := i["because"]; ok {
			e := "Body regex match: 0BC20225-2E72-4646-9202-8467972199E1"
			if b != e {
				t.Errorf("Expecting: %q, got: %q", e, b)
			}
		} else {
			t.Errorf("key not found: %q", "because")
		}
		// check exit
		if e, ok := i["exit"]; ok {
			if e.(float64) != 0 {
				t.Errorf("Expecting: 0 got: %v", e.(float64))
			}
		} else {
			t.Errorf("key not found: %q", "exit")
		}
		// check url
		if _, ok := i["url"]; !ok {
			t.Error("URL key not found")
		}
		wg.Done()
	}))
	defer log_s.Close()
	s := make(Services)
	re := regexp.MustCompile(`(?i)[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}`)
	s["s 1"] = Service{
		Name: "s 1",
		URL:  check_s.URL,
		Log:  log_s.URL,
		Expect: Expect{
			Body: *re,
		},
	}
	ez := &Epazote{
		Services: s,
	}
	wg.Add(1)
	ez.Supervice(s["s 1"])()
	wg.Wait()
}

func TestSuperviceBodyNoMatch(t *testing.T) {
	var wg sync.WaitGroup
	check_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		fmt.Fprintln(w, "Hello, epazote match 0BC20225-2E72-4646-9202-8467972199E1 regex")
	}))
	defer check_s.Close()
	log_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		decoder := json.NewDecoder(r.Body)
		var i map[string]interface{}
		err := decoder.Decode(&i)
		if err != nil {
			t.Error(err)
		}
		// check name
		if n, ok := i["name"]; ok {
			if n != "s 1" {
				t.Errorf("Expecting  %q, got: %q", "s 1", n)
			}
		} else {
			t.Errorf("key not found: %q", "name")
		}
		// check because
		if b, ok := i["because"]; ok {
			e := "Body no regex match: [a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}"
			if b != e {
				t.Errorf("Expecting: %q, got: %q", e, b)
			}
		} else {
			t.Errorf("key not found: %q", "because")
		}
		// check exit
		if e, ok := i["exit"]; ok {
			if e.(float64) != 1 {
				t.Errorf("Expecting: 1 got: %v", e.(float64))
			}
		} else {
			t.Errorf("key not found: %q", "exit")
		}
		// check output
		if o, ok := i["output"]; ok {
			e := "No defined cmd"
			if o != e {
				t.Errorf("Expecting %q, got %q", e, o)
			}
		} else {
			t.Errorf("key not found: %q", "output")
		}
		// check url
		if _, ok := i["url"]; !ok {
			t.Error("URL key not found")
		}
		wg.Done()
	}))
	defer log_s.Close()
	s := make(Services)
	re := regexp.MustCompile(`[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}`)
	s["s 1"] = Service{
		Name: "s 1",
		URL:  check_s.URL,
		Log:  log_s.URL,
		Expect: Expect{
			Body: *re,
		},
	}
	ez := &Epazote{
		Services: s,
	}
	wg.Add(1)
	ez.Supervice(s["s 1"])()
	wg.Wait()
}

func TestSuperviceNoGet(t *testing.T) {
	var wg sync.WaitGroup
	log_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		decoder := json.NewDecoder(r.Body)
		var i map[string]interface{}
		err := decoder.Decode(&i)
		if err != nil {
			t.Error(err)
		}
		// check name
		if n, ok := i["name"]; ok {
			if n != "s 1" {
				t.Errorf("Expecting  %q, got: %q", "s 1", n)
			}
		} else {
			t.Errorf("key not found: %q", "name")
		}
		// check because
		if b, ok := i["because"]; ok {
			e := "GET: Get http://: http: no Host in request URL"
			if b != e {
				t.Errorf("Expecting: %q, got: %q", e, b)
			}
		} else {
			t.Errorf("key not found: %q", "because")
		}
		// check exit
		if e, ok := i["exit"]; ok {
			if e.(float64) != 1 {
				t.Errorf("Expecting: 1 got: %v", e.(float64))
			}
		} else {
			t.Errorf("key not found: %q", "exit")
		}
		// check output
		if o, ok := i["output"]; ok {
			e := "exit status 1"
			if o != e {
				t.Errorf("Expecting %q, got %q", e, o)
			}
		} else {
			t.Errorf("key not found: %q", "output")
		}
		// check url
		if _, ok := i["url"]; !ok {
			t.Error("URL key not found")
		}
		wg.Done()
	}))
	defer log_s.Close()
	s := make(Services)
	s["s 1"] = Service{
		Name: "s 1",
		URL:  "http://",
		Log:  log_s.URL,
		Expect: Expect{
			Status: 200,
			IfNot: Action{
				Cmd: "test 1 -gt 2",
			},
		},
	}
	ez := &Epazote{
		Services: s,
	}
	wg.Add(1)
	ez.Supervice(s["s 1"])()
	wg.Wait()
}
func TestSuperviceNoGetStatus0(t *testing.T) {
	var wg sync.WaitGroup
	log_s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-agent") != "epazote" {
			t.Error("Expecting User-agent: epazote")
		}
		decoder := json.NewDecoder(r.Body)
		var i map[string]interface{}
		err := decoder.Decode(&i)
		if err != nil {
			t.Error(err)
		}
		// check name
		if n, ok := i["name"]; ok {
			if n != "s 1" {
				t.Errorf("Expecting  %q, got: %q", "s 1", n)
			}
		} else {
			t.Errorf("key not found: %q", "name")
		}
		// check because
		if b, ok := i["because"]; ok {
			e := "GET: Get http://: http: no Host in request URL"
			if b != e {
				t.Errorf("Expecting: %q, got: %q", e, b)
			}
		} else {
			t.Errorf("key not found: %q", "because")
		}
		// check exit
		if e, ok := i["exit"]; ok {
			if e.(float64) != 1 {
				t.Errorf("Expecting: 1 got: %v", e.(float64))
			}
		} else {
			t.Errorf("key not found: %q", "exit")
		}
		// check output
		if o, ok := i["output"]; ok {
			t.Errorf("key should not exist,content: %q", o)
		}
		// check url
		if _, ok := i["url"]; !ok {
			t.Error("URL key not found")
		}
		wg.Done()
	}))
	defer log_s.Close()
	s := make(Services)
	s["s 1"] = Service{
		Name: "s 1",
		URL:  "http://",
		Log:  log_s.URL,
		Expect: Expect{
			Status: 200,
			IfNot: Action{
				Cmd: "test 3 -gt 2",
			},
		},
	}
	ez := &Epazote{
		Services: s,
	}
	wg.Add(1)
	ez.Supervice(s["s 1"])()
	wg.Wait()
}
