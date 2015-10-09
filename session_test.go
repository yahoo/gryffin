package gryffin

import (
	"sync"
	"testing"
	"time"
)

func TestNewGryffinStore(t *testing.T) {

	t.Parallel()

	store1 := NewGryffinStore(true)
	store2 := NewGryffinStore(true)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		store1.See("foo", "oracle", uint64(0x1234))
		b := <-store1.GetSndChan()
		t.Log("Store1 got ", string(b))
		store2.GetRcvChan() <- b
		wg.Done()
	}()

	wg.Wait()
	for i := 0; i < 100000; i++ {
		if store2.Seen("foo", "oracle", uint64(0x1234), 2) {
			t.Logf("Store2 see the new value in %d microseconds.", i)
			break
		}
		time.Sleep(1 * time.Microsecond)
	}

	if !store2.Seen("foo", "oracle", uint64(0x1234), 2) {
		t.Error("2nd store should see the value in oracle.", store2.Oracles)
	}

}
