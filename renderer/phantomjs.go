// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renderer

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/yahoo/gryffin"
	_ "github.com/yahoo/gryffin/renderer/resource"
)

/* all of these are the JSON struct for phantomjs render.js */

type PhantomJSRenderer struct {
	BaseRenderer
	Timeout int
	process *os.Process
}

type input struct {
	Method         string       `json:"method"`
	AllowedDomains []string     `json:"allowed_domains,omitempty"`
	Headers        inputHeaders `json:"headers"`
}

type inputHeaders struct {
	AcceptEncoding string `json:"Accept-Encoding"`
	AcceptLanguage string `json:"Accept-Language"`
	Cookie         string
	UserAgent      string `json:"User-Agent"`
}

type details struct {
	Links        []link
	Forms        []form
	ChildFrames  []link
	SubResources []link
	Redirects    []link
	MainFrame    []link
}

type link struct {
	Text string
	Url  string
}

type form struct {
	Data     string
	DataType string
	Method   string
	Url      string
}

type response struct {
	Headers     map[string][]string
	Body        string
	ContentType string
	Status      int
	Url         string
	Details     details
}

type responseMessage struct {
	Response response
	Elapsed  int
	Ok       int
}

type domMessage struct {
	Action   string
	Events   []string
	KeyChain []string
	JSError  []string
}

type message struct {
	*responseMessage
	*domMessage
	*details
	Signature string
	MsgType   string
}

type noCloseReader struct {
	io.Reader
}

func (r noCloseReader) Close() error {
	return nil
}

func (m *response) fill(s *gryffin.Scan) {

	/*
	   {"response":{"headers":{"Date":["Thu, 30 Jul 2015 00:13:43 GMT"],"Set-Cookie":["B=82j3nrdarir1n&b=3&s=23; expires=Sun, 30-Jul-2017 00:13:43 GMT; path=/; domain=.yahoo.com"]

	*/
	resp := &http.Response{
		Request:    s.Request,
		StatusCode: m.Status,
		Status:     strconv.FormatInt(int64(m.Status), 10),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     m.Headers,
		Body:       noCloseReader{strings.NewReader(m.Body)},
	}

	s.Response = resp
	s.ReadResponseBody()

}

func (f *form) toScan(parent *gryffin.Scan) *gryffin.Scan {
	m := strings.ToUpper(f.Method)
	u := f.Url
	var r io.Reader
	if m == "POST" {
		r = ioutil.NopCloser(strings.NewReader(f.Data))
	} else {
		parsed, err := url.Parse(u)
		if err == nil {
			parsed.RawQuery = f.Data
			u = parsed.String()
		}
	}

	if req, err := http.NewRequest(m, u, r); err == nil {
		s := parent.Spawn()
		s.MergeRequest(req)
		return s
	}
	// invalid url
	return nil
}

func (l *link) toScan(parent *gryffin.Scan) *gryffin.Scan {
	if req, err := http.NewRequest("GET", l.Url, nil); err == nil {
		s := parent.Spawn()
		s.MergeRequest(req)
		return s
	}
	// invalid url
	return nil
}

func (r *PhantomJSRenderer) extract(stdout io.ReadCloser, s *gryffin.Scan) {

	defer close(r.done)

	dec := json.NewDecoder(stdout)
	for {
		var m message
		err := dec.Decode(&m)
		if err == io.EOF {
			return
		} else {
			if m.responseMessage != nil {
				m.Response.fill(s)
				if s.IsDuplicatedPage() {
					return
				}
				r.chanResponse <- s
				r.parseDetails(&m.Response.Details, s)
			}

			if m.details != nil {
				r.parseDetails(m.details, s)
			}
		}
	}
}

func (r *PhantomJSRenderer) parseDetails(d *details, s *gryffin.Scan) {
	v := reflect.ValueOf(*d)
	for i := 0; i < v.NumField(); i++ {
		if links, ok := v.Field(i).Interface().([]link); ok {
			for _, link := range links {
				if newScan := link.toScan(s); newScan != nil && newScan.IsScanAllowed() {
					r.chanLinks <- newScan
				}
			}
		}
		if forms, ok := v.Field(i).Interface().([]form); ok {
			for _, form := range forms {
				if newScan := form.toScan(s); newScan != nil && newScan.IsScanAllowed() {
					r.chanLinks <- newScan
				}
			}
		}
	}
}

func (r *PhantomJSRenderer) kill(reason string, s *gryffin.Scan) {
	if err := r.process.Kill(); err == nil {
		s.Logmf("PhantomjsRenderer.Do", "[%s] Terminating the crawl process.", reason)
	}
}

func (r *PhantomJSRenderer) wait(s *gryffin.Scan) {

	select {
	case <-r.done:
		r.kill("Cleanup", s)
	case <-time.After(time.Duration(r.Timeout) * time.Second):
		r.kill("Timeout", s)
	}
	close(r.chanResponse)
	close(r.chanLinks)
}

func (r *PhantomJSRenderer) Do(s *gryffin.Scan) {

	r.chanResponse = make(chan *gryffin.Scan, 10)
	r.chanLinks = make(chan *gryffin.Scan, 10)
	r.done = make(chan string)

	// Construct the command.
	// render.js http(s)://<host>[:port][/path] [{"method":"post", "data":"a=1&b=2"}]
	url := s.Request.URL.String()
	cookies := make([]string, 0)
	// ua := s.Request.UserAgent()
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36"

	for _, c := range s.Cookies {
		cookies = append(cookies, c.String())
	}

	arg := input{
		Method: s.Request.Method,
		Headers: inputHeaders{
			UserAgent: ua,
			Cookie:    strings.Join(cookies, ";"),
		},
	}

	opt, err := json.Marshal(arg)
	if err != nil {
		s.Error("PhantomjsRenderer.Do", err)
		return
	}

	// s.Logmf("PhantomjsRenderer.Do", "Running: render.js %s '%s'", url, string(opt))
	s.Logmf("PhantomjsRenderer.Do", "Running: render.js")

	cmd := exec.Command(
		"phantomjs",
		"--ssl-protocol=any",
		"--ignore-ssl-errors=true",
		os.Getenv("GOPATH")+"/src/github.com/yahoo/gryffin/renderer/resource/render.js",
		url,
		string(opt))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.Error("PhantomjsRenderer.Do", err)
		return
	}

	if err := cmd.Start(); err != nil {
		s.Error("PhantomjsRenderer.Do", err)
		return
	}

	r.process = cmd.Process

	// wait until done or timeout.
	go r.extract(stdout, s)
	go r.wait(s)

	// cmd.Wait will close the stdout pipe.
	go cmd.Wait()

}
