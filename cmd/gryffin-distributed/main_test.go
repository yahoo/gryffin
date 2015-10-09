// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Unit test for gryffin-distributed is still on todo list.
//
// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"

// 	"github.com/yahoo/gryffin"
// )

// var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Hello World"))
// })

// var ts = httptest.NewServer(handler)

// func TestMain(t *testing.T) {
// 	if os.Getenv("INTEGRATION") == "" {
// 		t.Skip("Skip integration tests.")
// 	}
// 	scan := gryffin.NewScan("GET", ts.URL, "")
// 	linkChannels(scan)

// }
