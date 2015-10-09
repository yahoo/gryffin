// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arachni

import (
	"os"
	"testing"

	"github.com/yahoo/gryffin"
)

func TestFuzzer(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("Skip integration tests.")
	}
	s := &Fuzzer{}
	scan := gryffin.NewScan("GET", "http://127.0.0.1:8081/xss/reflect/full1?in=change_me", "")
	c, err := s.Fuzz(scan)
	if err != nil {
		t.Error(err)
	}
	if c == 0 {
		t.Error("No issue detected.")
	}
}
