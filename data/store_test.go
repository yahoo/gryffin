// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package data

import (
	"testing"
)

func testStore(t *testing.T, s Store) {
	s.Set("hello", "world")
	if v, ok := s.Get("hello"); !ok || v != "world" {
		t.Error("Get and Set is inconsistent.", v)
	}

	s.Set("foo", 100)
	if n := s.IncrBy("foo", 10); n != 110 {
		t.Error("Incr failed.")
	}
	if v, ok := s.Get("foo"); v.(int64) != 110 {
		t.Errorf("Incr is inconsistent %t, %t and %s", ok, v.(int64) == 110, v)
	}

}

func benchStore(b *testing.B, s Store) {
	s.Set("hello", "world")
	s.Set("foo", 100)
	s.IncrBy("foo", 10)
}
