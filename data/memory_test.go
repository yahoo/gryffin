// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package data

import (
	"testing"
)

func TestMemoryStore(t *testing.T) {
	t.Parallel()
	testStore(t, NewMemoryStore())
}

func BenchmarkMemoryStore(b *testing.B) {
	s := NewMemoryStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStore(b, s)
	}

}
