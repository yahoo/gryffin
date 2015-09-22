// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yahoo/gryffin/data"
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
	s := NewScan("GET", ts.URL, "", nil, nil)
	if s == nil {
		t.Error("Scan is nil.")
	}
	t.Logf("Jobid: %s.", s.Job.ID)
}

func TestPoke(t *testing.T) {
	t.Parallel()
	client := &http.Client{}
	s := NewScan("GET", ts.URL, "", nil, nil)
	err := s.Poke(client)
	if err != nil {
		t.Error(err)
	}
}

// func TestPokeNonExist(t *testing.T) {
// 	t.Parallel()
// 	client := &http.Client{}
// 	s := NewScan("GET", ts.URL, "", nil, nil)
// 	err := s.Poke(client)
// 	if err == nil {
// 		t.Error("Expected error fetching non-existing host.")
// 	}
// }

func TestSpawn(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "", nil, nil)
	s2 := s.Spawn()
	if s.Request.URL != s2.Request.URL {
		t.Error("Spawn gives a request with different URL.")
	}
}

func TestMergeRequest(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "", nil, nil)
	s.Request.Header.Set("User-Agent", "foo")
	r, _ := http.NewRequest("GET", ts.URL, nil)
	s.MergeRequest(r)

	if s.Request.UserAgent() != "foo" {
		t.Errorf("Merge request got a different user agent: %s", s.Request.UserAgent())
	}

}

func TestMergeRequestRelative(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "", nil, nil)
	s.Request.Header.Set("User-Agent", "foo")
	r, _ := http.NewRequest("GET", "/#", nil)
	s.MergeRequest(r)

	if s.Request.URL.String() != ts.URL+"/" {
		t.Errorf("Merge request cannot resolve relative url: %s", s.Request.URL)
	}
}

func TestRateLimit(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "", data.NewMemoryStore(), nil)
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

func TestLog(t *testing.T) {
	t.Parallel()
	s := NewScan("GET", ts.URL, "", nil, os.Stderr)
	s.Logf("test logf %s", "foo")
	s.Log(s)
	s.Logm("test service", "foo message")
}
