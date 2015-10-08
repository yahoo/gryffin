// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

var h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
})

var ts = httptest.NewServer(h)

func TestGenRandomID(t *testing.T) {
	t.Parallel()
	id := GenRandomID()
	if len(id) == 0 {
		t.Error("Empty ID from GenRandomID.")
	}
}

func TestNewScan(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	if s == nil {
		t.Error("Scan is nil.")
	}
	// TODO - verify s.DomainAllowed.
	t.Logf("Jobid: %s.", s.Job.ID)
}

func TestNewScanInvalid(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", "%a", "")
	if s != nil {
		t.Error("Scan is not nil with invalid URL.", s.Request)
	}
}

func TestNewScanFromJson(t *testing.T) {
	t.Parallel()

	// Test arbritary url.
	s := NewScan("GET", ts.URL, "")
	_ = s.Poke(&http.Client{})
	j := s.Json()

	if j == nil {
		t.Error("scan.Json should return a json string.")
	}

	s2 := NewScanFromJson(j)
	if s2 == nil {
		t.Error("NewScanFromJson should return a scan.")
	}
	t.Log(s2)

}

func TestGetOrigin(t *testing.T) {
	t.Parallel()
	u, _ := url.Parse("http://127.0.0.1:1234/foo/bar?")
	o := getOrigin(u)
	if o != "http://127.0.0.1:1234" {
		t.Error("getOrigin is not valid", u, o)
	}
}

func TestScanPoke(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	err := s.Poke(&http.Client{})
	if err != nil {
		t.Error(err)
	}
}

func TestScanPokeInvalidURL(t *testing.T) {
	t.Parallel()
	client := &http.Client{}
	s := NewScan("GET", "/foo", "")
	err := s.Poke(client)
	if err == nil {
		t.Error("Expect an error with invalid scheme.")
	}
	t.Log("Negative test: Invalid url got ", err)
}

func TestScanSpawn(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	s.Poke(&http.Client{})
	s2 := s.Spawn()
	if s.Request.URL != s2.Request.URL {
		t.Error("Spawn gives a request with different URL.")
	}
}

func TestScanMergeRequest(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "foo=bar")
	s.Poke(&http.Client{})
	s.Request.Header.Set("User-Agent", "foo")
	s.Cookies = []*http.Cookie{
		&http.Cookie{Name: "cookie-name-1", Value: "cookie-value-1"},
	}

	r, _ := http.NewRequest("GET", ts.URL, strings.NewReader("quz=quxx"))
	s.MergeRequest(r)
	if s.Request.UserAgent() != "foo" {
		t.Errorf("Merge request got a different user agent: %s", s.Request.UserAgent())
	}
}

func TestScanMergeRequestRelative(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	s.Request.Header.Set("User-Agent", "foo")
	r, _ := http.NewRequest("GET", "/#", nil)
	s.MergeRequest(r)

	if s.Request.URL.String() != ts.URL+"/" {
		t.Errorf("Merge request cannot resolve relative url: %s", s.Request.URL)
	}
}

func TestScanReadResponseBody(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	s.Poke(&http.Client{})
	s.ReadResponseBody()
	if s.ResponseBody == "" {
		t.Error("Empty ResponseBody")
	}
	// t.Log(s.ResponseBody)
}

func TestScanUpdateFingerprint(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", "http://127.0.0.1", "")
	s.UpdateFingerprint()
	if !reflect.DeepEqual(
		s.Fingerprint,
		Fingerprint{0x7233A9A31DEADAF2, 0x7233A9A31DEADAF2, 0xF8A4322BD612093C, 0, 0}) {
		t.Error("Fingerprint mismatch", s.Fingerprint)
	}
}

func TestScanResponseFingerprint(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	s.Poke(&http.Client{})
	s.UpdateFingerprint()
	if s.Fingerprint.ResponseSimilarity != 0x62C1D0803B2AB139 {
		t.Error("Fingerprint mismatch", s.Fingerprint)
	}
}

func TestScanRateLimit(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "")
	for i := 0; i < 5; i++ {
		d := s.RateLimit()
		if d > 0 {
			t.Errorf("Got delayed for %d", d)
		}
	}
	d := s.RateLimit()
	if d == 0 {
		t.Errorf("No delay after 5 request. Got %d", d)
	}
}

func TestScanIsScanAllowed(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", "http://foo.com", "")

	r, _ := http.NewRequest("GET", "http://bar.com", nil)
	s.MergeRequest(r)
	if s.IsScanAllowed() {
		t.Error("IsScanAllowed should return false", s)
	}

	r, _ = http.NewRequest("GET", "http://foo.com/test", nil)
	s.MergeRequest(r)
	if !s.IsScanAllowed() {
		t.Error("IsScanAllowed should return true", s)
	}

	s2 := NewScan("GET", "/no-domain", "")
	if !s2.IsScanAllowed() {
		t.Error("IsScanAllowed should return true", s2.Request.URL)
	}
}

func TestScanCrawlAsync(t *testing.T) {
	// TODO ...
	t.Parallel()
}

func TestScanIsDuplicatedPage(t *testing.T) {
	t.Parallel()
	s1 := NewScan("GET", ts.URL, "")
	_ = s1.Poke(&http.Client{})
	if s1.IsDuplicatedPage() {
		t.Error("IsDuplicatedPage should return false for the first page", s1)
	}

	s2 := s1.Spawn()
	r, _ := http.NewRequest("GET", ts.URL, nil)
	s2.MergeRequest(r)
	_ = s2.Poke(&http.Client{})
	if !s2.IsDuplicatedPage() {
		t.Errorf("IsDuplicatedPage should return true for the second page with same Job ID.\n1st Page: %064b\n2nd Page: %064b\n",
			s1.Fingerprint.ResponseSimilarity, s2.Fingerprint.ResponseSimilarity)
	}

	s3 := Scan(*s1)
	s3.Job.ID = "ABCDEF123456"
	if s3.IsDuplicatedPage() {
		t.Error("IsDuplicatedPage should return false for the a page with new Job ID", s3)
	}

}

func TestScanFuzz(t *testing.T) {
	// TODO ...
	t.Parallel()
}

func TestScanShouldCrawl(t *testing.T) {
	t.Parallel()
	s1 := NewScan("GET", ts.URL, "")
	if !s1.ShouldCrawl() {
		t.Error("ShouldCrawl should return true for the first page", s1)
	}

	s2 := s1.Spawn()
	r, _ := http.NewRequest("GET", ts.URL, nil)
	s2.MergeRequest(r)

	if s2.ShouldCrawl() {
		t.Errorf("ShouldCrawl should return false for the second page with same Job ID.\n1st Page: %064b\n2nd Page: %064b\n",
			s1.Fingerprint.ResponseSimilarity, s2.Fingerprint.ResponseSimilarity)
	}

	s3 := Scan(*s1)
	s3.Job.ID = "ABCDEF123456"
	if !s3.ShouldCrawl() {
		t.Error("ShouldCrawl should return true for the a page with new Job ID", s3)
	}
}

func TestScanLog(t *testing.T) {
	t.Parallel()
	SetLogWriter(os.Stdout)
	s := NewScan("GET", ts.URL, "")
	s.Log(s)
}
