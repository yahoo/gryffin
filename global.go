// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"io"
	"sync"
)

var (
	memoryStore   *GryffinStore
	logWriter     io.Writer
	memoryStoreMu sync.Mutex
	logWriterMu   sync.Mutex
)

// SetMemoryStore sets the package internal global variable
// for the memory store.
func SetMemoryStore(m *GryffinStore) {
	memoryStoreMu.Lock()
	memoryStore = m
	memoryStoreMu.Unlock()
}

// SetLogWriter sets the log writer.
func SetLogWriter(w io.Writer) {
	logWriterMu.Lock()
	logWriter = w
	logWriterMu.Unlock()
}
