// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package data

import (
	"testing"
)

func TestRedisStore(t *testing.T) {
	t.Parallel()
	testStore(t, NewRedisStore("redis://localhost:6379/0"))
}

func BenchmarkRedisStore(b *testing.B) {
	s := NewRedisStore("redis://localhost:6379/0")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStore(b, s)
	}

}
