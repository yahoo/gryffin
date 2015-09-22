// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gryffin is an application scanning infrastructure.
*/
package gryffin

import (
	"bytes"
	// "fmt"
	// "encoding/hex"
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	// "log"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
	// "os"
	// "strings"

	"github.com/yahoo/gryffin/data"
	"github.com/yahoo/gryffin/html-distance"
)

// TODO, add https://godoc.org/github.com/mfonda/simhash
// TODO, add it via vendoring.

// A Scan consists of the job, target, request and response.
type Scan struct {
	// ID is a random ID to identify this particular scan.
	// if ID is empty, this scan should not be performed (but record for rate limiting).
	ID           string
	Job          *Job
	Request      *http.Request
	RequestBody  string
	Response     *http.Response
	ResponseBody string
	Cookies      []*http.Cookie
	Fingerprint  Fingerprint
	HitCount     int
	Session      data.Store
	LogWriter    io.Writer
}

// Job stores the job id and config (if any).
type Job struct {
	ID             string
	DomainsAllowed []string // Domains that we would crawl
	// Config         Config   // unused currently. It would store job meta data.
}

// type Config map[string]string

// Fingerprint contains all the different types of hash for the Scan (Request & Response)
type Fingerprint struct {
	Origin             uint64 // origin
	URL                uint64 // origin + path
	Request            uint64 // method, url, body
	RequestFull        uint64 // request + header
	ResponseSimilarity uint64
	//  TODO - add simhash here.
}

// // Issue stores the findings from the fuzzer.
// type Issue struct {
// 	// TODO - implement this
// 	// *Job
// 	// *WebSite
// }

// HTTPDoer interface is to be implemented by http.Client
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Fuzzer runs the fuzzing.
type Fuzzer interface {
	Fuzz(*Scan) (int, error)
}

// Renderer is an interface for implementation HTML DOM renderer and obtain the response body and links.
// Since DOM construction is very likely to be asynchronous, we return the channels to receive response and links.
type Renderer interface {
	Do(*Scan)
	GetRequestBody() <-chan *Scan
	GetLinks() <-chan *Scan
}

// // Storage is an interface for storing sessions and persistent objects
// type Storage interface {
// 	Get(key string) (value interface{}, ok bool)
// 	Set(key string, value interface{})
// }

// LogMessage contains the data fields to be marshall as a json for forwarding to the log processor.
type LogMessage struct {
	Service string
	Msg     string
	// Fingerprint Fingerprint
	Method string
	Url    string
}

// // Logger is an interface for writing logs.
// // In full deployment mode, logger may send events to other receivers.
// type Logger interface {
// 	Log(...interface{})
// 	Logf(string, ...interface{})
// }

// // StdLogger is a simple logger that writes to stdout.
// type StdLogger struct{}

// // Log calls log.Println
// func (s *StdLogger) Log(v ...interface{}) {
// 	log.Println(v...)
// }

// // Logf calls log.Printf
// func (s *StdLogger) Logf(format string, v ...interface{}) {
// 	log.Printf(format, v...)
// }

// TODO
// similarity oracle. Will move these to interface and let the main package create it.
var oracle = distance.NewOracle()

// NewScan creates a scan.
func NewScan(method, url, post string, session data.Store, writer io.Writer) *Scan {

	id := GenRandomID()

	job := &Job{ID: GenRandomID()}

	req, err := http.NewRequest(method, url, ioutil.NopCloser(strings.NewReader(post)))
	if err != nil {
		// s.Log("Invalid url for NewScan: %s", err)
		return nil
	}

	// put the host component of the url as the domains to be allowed
	job.DomainsAllowed = []string{strings.SplitN(req.URL.Host, ":", 2)[0]}

	// // Add chrome user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.107 Safari/537.36")

	return &Scan{
		ID:          id,
		Job:         job,
		Request:     req,
		RequestBody: post,
		Session:     session,
		LogWriter:   writer,
	}
}

func NewScanFromJson(b []byte, base *Scan) *Scan {
	var scan Scan
	json.Unmarshal(b, &scan)
	if base != nil {
		scan.LogWriter = base.LogWriter
		scan.Session = base.Session
	}
	// fmt.Printf("Cookie: %v\n", scan.Cookies)
	return &scan
}

// getOrigin returns the Origin of the URL (scheme, hostname, port )
func getOrigin(u *url.URL) string {
	return u.Scheme + u.Host
}

// MergeRequest merge the request field in scan with the existing one.
func (s *Scan) MergeRequest(req *http.Request) {

	// set cookie from response (if it is not done..)
	if s.Response != nil {
		s.Cookies = append(s.Cookies, s.Response.Cookies()...)
		// s.CookieJar.SetCookies(s.Request.URL, s.Response.Cookies())
	}

	// read the request body, and then reset the reader
	var post []byte
	if req.Body != nil {
		if post, err := ioutil.ReadAll(req.Body); err != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(post))
		}
	}

	// resolve relative url.
	if !req.URL.IsAbs() {
		req.URL = s.Request.URL.ResolveReference(req.URL)
	}

	// TODO - drop if Method, URL, Body are same..
	if req == s.Request {
		// s.Logf("Result after merge generate same request.", nil)
	}

	// swap
	prevReq := s.Request
	s.Request = req
	s.RequestBody = string(post)

	// TODO - handle relative URL .

	// Create a cookie jar, add cookie list (so cookie jar reject invalid cookie.)
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(req.URL, s.Cookies)

	// reset cookies
	s.Cookies = make([]*http.Cookie, 0)
	for _, c := range jar.Cookies(req.URL) {
		req.AddCookie(c)
		s.Cookies = append(s.Cookies, c)
	}

	// Add user agent
	req.Header.Set("User-Agent", prevReq.UserAgent())

	// Add referrer - TODO, perhaps we don't need this!

	// remove Response.
	s.Response = nil
	s.ResponseBody = ""

}

// Spawn spawns a new scan object with a different ID.
func (s *Scan) Spawn() *Scan {
	id := GenRandomID()
	job := *s.Job
	req := *s.Request // copy the value.

	post := s.RequestBody
	s.Request.Body = ioutil.NopCloser(strings.NewReader(post))

	// get the cookiejar, save the new cookies
	// jar := s.CookieJar
	cookies := s.Cookies[:]
	if s.Response != nil {
		cookies = append(cookies, s.Response.Cookies()...)
		// jar.SetCookies(s.Request.URL, s.Response.Cookies())
	}

	return &Scan{
		ID:          id,
		Job:         &job,
		Request:     &req,
		RequestBody: post,
		Cookies:     cookies,
		Session:     s.Session,
		LogWriter:   s.LogWriter,
	}
}

// Poke checks if the target is up.
func (s *Scan) Poke(client HTTPDoer) (err error) {

	s.Logm("Poke", "Poking")

	// Add 5s timeout if it is http.client
	switch client.(type) {
	case *http.Client:
		client.(*http.Client).Timeout = time.Duration(3) * time.Second
	}

	// delete the similarity case for the domain.
	s.Session.DelPrefix("hash/unique/" + s.Request.URL.Host)

	// http.Request is embeded in a Request embeded in a Scan.
	s.Response, err = client.Do(s.Request)
	if err != nil {
		s.Logm("Poke", "Failed")
		return
	}

	s.ReadResponseBody()

	s.HitCount++
	return
}

/*
// LogWriter writes log.
type LogWriter struct{}

// Write sends JSON message to the logger machine.
func (w *LogWriter) Write(p []byte) (n int, err error) {
	n, err = os.Stdout.Write(p)
	// n = len(p)
	return
}

// Logger is the default logger for this package.
var Logger = log.New(&LogWriter{}, "", log.LstdFlags)
*/

// ReadResponseBody read Response.Body and fill it to ReadResponseBody.
// It will also reconstruct the io.ReaderCloser stream.
func (s *Scan) ReadResponseBody() {
	if s.ResponseBody == "" && s.Response != nil {
		if b, err := ioutil.ReadAll(s.Response.Body); err == nil {
			s.ResponseBody = string(b)
			s.Response.Body = ioutil.NopCloser(bytes.NewReader(b))
		}
	}
}

func hash(s string) uint64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return h.Sum64()
}

// UpdateFingerprint updates the fingerprint field.
func (s *Scan) UpdateFingerprint() {
	f := &s.Fingerprint
	if s.Request != nil {
		if f.Origin == 0 {
			f.Origin = hash(getOrigin(s.Request.URL))
		}
		if f.URL == 0 {
			f.URL = hash(s.Request.URL.String())
		}
		if f.Request == 0 {
			f.Request = hash(s.Request.URL.String() + "\n" + s.RequestBody)
		}
		if f.RequestFull == 0 {
			// TODO
		}
	}

	if f.ResponseSimilarity == 0 {
		if r := strings.NewReader(s.ResponseBody); s.ResponseBody != "" && r != nil {
			f.ResponseSimilarity = distance.Fingerprint(r, 3)
			s.Logm("Fingerprint", "Computed")
		}
	}

}

// RateLimit checks whether we are under the allowed rate for crawling the site.
// It returns a delay time to wait to check for ReadyToCrawl again.
func (s *Scan) RateLimit() int {
	store := s.Session
	// for each 5 second epoch, we create a key and see how many crawls are done.
	ts := time.Now().Truncate(5 * time.Second).Unix()
	k := "rate/" + s.Request.URL.Host + "/" + strconv.FormatInt(ts, 10)
	if v, ok := store.Get(k); ok {
		if v.(int64) >= 5 {
			// s.Logm("RateLimit", "Delay 5 second")
			// s.Logf("Wait for 5 second for %s (v:%d)", s.Request.URL, v)
			return 5
		}
		// ready to crawl.
		// TODO - this is not atomic.
		c, _ := store.Get(k)
		store.Set(k, c.(int64)+1)
		// s.Logm("RateLimit", "No Delay")
		return 0
	}

	store.Set(k, 1)
	// s.Logm("RateLimit", "No Delay")
	return 0
}

// IsScanAllowed check if the request URL is allowed per Job.DomainsAllowed.
func (s *Scan) IsScanAllowed() bool {
	// relative URL
	if !s.Request.URL.IsAbs() {
		return true
	}

	host := strings.SplitN(s.Request.URL.Host, ":", 2)[0]
	for _, allowed := range s.Job.DomainsAllowed {
		if host == allowed {
			return true
		}
	}
	return false
}

// CrawlAsync run the crawling asynchronously.
func (s *Scan) CrawlAsync(r Renderer) {
	s.Logm("CrawlAsync", "Started")
	if s.IsScanAllowed() {
		r.Do(s)
	} else {
		s.Logm("CrawlAsync", "Scan Not Allowed")
	}
}

// IsDuplicatedPage checks if we should proceed based on the Response
func (s *Scan) IsDuplicatedPage() bool {

	s.UpdateFingerprint()
	store := s.Session

	f := s.Fingerprint.ResponseSimilarity
	// TODO - tune the proximity score...
	if !oracle.Seen(f, 2) {
		k := "hash/unique/" + s.Request.URL.Host + "/" + strconv.FormatUint(s.Fingerprint.ResponseSimilarity, 16)
		store.Set(k, true)
		oracle.See(f)
		s.Logm("IsDuplicatedPage", "Unique Page")
		return false
	}

	return true
}

// IsUnique checks if the response is unique or not.
func (s *Scan) IsUnique() bool {
	s.UpdateFingerprint()
	store := s.Session

	k := "hash/unique/" + s.Request.URL.Host + "/" + strconv.FormatUint(s.Fingerprint.ResponseSimilarity, 16)

	if _, ok := store.Get(k); ok {
		// s.Logger.Logf("Uniqueness is %t, %s", true, s.Request.URL)
		return true
	}

	// s.Logger.Logf("Uniqueness is %t, %s", false, s.Request.URL)
	return false
}

// Scan runs the vulnerability fuzzer, return the issue count
func (s *Scan) Fuzz(fuzzer Fuzzer) (int, error) {
	c, err := fuzzer.Fuzz(s)
	return c, err
}

// // ExtractLinks extracts the list of links found from the responseText in the Scan.
// func (s *Scan) ExtractLinks() (scans []Scan, err error) {

// 	return
// }

// ApplyLinkRules checks if the links should be queued for next crawl.
func (s *Scan) ApplyLinkRules() bool {

	s.UpdateFingerprint()
	store := s.Session

	// k := "hash/" + hex.EncodeToString(s.Fingerprint.URL)
	k := "hash/url/" + strconv.FormatUint(s.Fingerprint.URL, 16)

	// link seen before.
	if _, ok := store.Get(k); ok {
		// s.Logm("LinkRules", "Duplicated")
		return false
	}
	store.Set(k, true)
	return true
}

func (s *Scan) Json() []byte {
	b, _ := json.Marshal(s)
	return b

}

// TODO - LogFmt (fmt string)
// TODO - LogI (interface)
func (s *Scan) Error(service string, err error) {
	errmsg := fmt.Sprint(err)
	s.Logm(service, errmsg)
}

func (s *Scan) Logmf(service, format string, a ...interface{}) {
	s.Logm(service, fmt.Sprintf(format, a...))
}

// Logm sends a LogMessage to Log processor.
func (s *Scan) Logm(service, msg string) {
	// TODO - improve the efficiency of this.
	m := &LogMessage{
		Service: service,
		Msg:     msg,
		// Fingerprint: s.Fingerprint,
		Method: s.Request.Method,
		Url:    s.Request.URL.String(),
	}
	s.Log(m)
}

func (s *Scan) Logf(format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	s.Log(str)
}

func (s *Scan) Log(v interface{}) {
	if s.LogWriter == nil {
		return
	}
	encoder := json.NewEncoder(s.LogWriter)
	encoder.Encode(v)
}
