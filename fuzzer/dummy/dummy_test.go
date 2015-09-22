// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dummy

import (
	"os"
	"testing"

	"github.com/yahoo/gryffin"
)

func TestFuzzer(t *testing.T) {

	f := &Fuzzer{}
	scan := gryffin.NewScan("GET", "http://www.yahoo.com", "", nil, os.Stdout)
	_, err := f.Fuzz(scan)
	if err != nil {
		t.Error(err)
	}

}
