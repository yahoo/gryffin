// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package distance

import (
	"testing"
)

func TestNewOracle(t *testing.T) {
	// just add 0 and 1.
	oracle := NewOracle()
	for i := uint64(1); i < 2; i++ {
		oracle.See(i)
	}
	r := uint8(2)
	for i := uint64(0); i < 30; i++ {
		t.Logf("Has the oracle seen anything closed to %02d (%08b) within distance of %d? %t", i, i, r, oracle.Seen(i, r))
	}

}

func BenchmarkOracleSee(b *testing.B) {
	oracle := NewOracle()
	for i := 0; i < b.N; i++ {
		// for i := uint64(1); i < 10000; i++ {
		oracle.See(uint64(i))
		// }
	}
}

func BenchmarkOracleSeen(b *testing.B) {
	oracle := NewOracle()
	for i := uint64(1); i < 1000000; i++ {
		oracle.See(i)
	}
	b.ResetTimer()
	r := uint8(2)
	for i := 0; i < b.N; i++ {
		oracle.Seen(uint64(i), r)
	}
}
