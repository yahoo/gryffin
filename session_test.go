package gryffin

import (
	"sync"
	"testing"
	"time"
)

func TestNewGryffinStore(t *testing.T) {

	t.Parallel()

	store1 := NewSharedGryffinStore()
	store2 := NewSharedGryffinStore()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		store1.See("foo", "oracle", uint64(0x1234))
		var b []byte
		b = <-store1.GetSndChan()
		t.Log("Store1 got ", string(b))
		store2.GetRcvChan() <- b

		store1.See("foo", "hash", uint64(0x5678))
		b = <-store1.GetSndChan()
		t.Log("Store1 got ", string(b))
		store2.GetRcvChan() <- b
		wg.Done()
	}()

	wg.Wait()
	for i := 0; i < 100000; i++ {
		if store2.Seen("foo", "oracle", uint64(0x1234), 2) {
			t.Logf("Store2 see the new oracle value in %d microseconds.", i)
			break
		}
		time.Sleep(1 * time.Microsecond)
	}

	if !store2.Seen("foo", "oracle", uint64(0x1234), 2) {
		t.Error("2nd store should see the oracle value in oracle.", store2.Oracles)
	}

	for i := 0; i < 100000; i++ {
		if store2.Seen("foo", "hash", uint64(0x5678), 2) {
			t.Logf("Store2 see the new hash value in %d microseconds.", i)
			break
		}
		time.Sleep(1 * time.Microsecond)
	}

	if !store2.Seen("foo", "hash", uint64(0x5678), 2) {
		t.Error("2nd store should see the hash value in hashes.", store2.Hashes)
	}
}
