// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renderer

import (
	"os"
	"testing"

	"github.com/yahoo/gryffin"
)

func testCrawlAsync(t *testing.T, r gryffin.Renderer) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("Skip integration tests.")
	}

	url := "https://www.yahoo.com/"

	s := gryffin.NewScan("GET", url, "")
	r.Do(s)
	s = <-r.GetRequestBody()
	// t.Logf("Got async body %s", s)
	for link := range r.GetLinks() {
		t.Logf("Got link %s", link.Request.URL)
	}
}
