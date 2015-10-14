// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"io"
	// "io/ioutil"
)

var memoryStore *GryffinStore
var logWriter io.Writer

func SetMemoryStore(m *GryffinStore) {
	memoryStore = m
}

func SetLogWriter(w io.Writer) {
	logWriter = w
}
