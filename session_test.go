package gryffin

import (
	"testing"
)

func TestNewGryffinStore(t *testing.T) {

	chanUpdate := make(chan []byte)

	store := NewGryffinStore(chanUpdate)
	_ = store
	store.See("foo", "oracle", uint64(0x1234))
	select {
	case b := <-chanUpdate:
		t.Log(string(b))
	default:
		t.Log("Got Nothing.")

	}
}

// func TestSharedCache(t *testing.T) {

//  i1 := make(chan []byte, 10)
//  o1 := make(chan []byte, 10)
//  s1 := NewGryffinStore(i1, o1)

//  i2 := make(chan []byte, 10)
//  o2 := make(chan []byte, 10)
//  s2 := NewGryffinStore(i2, o2)

//  s1.See("testing", uint64(0x1234))

//  msg := <-o1
//  t.Log("o1", string(msg))
//  fmt.Println("Send message to i2", string(msg))
//  i2 <- msg

//  time.Sleep(1 * time.Second)

//  t.Log(s1.Oracles)
//  t.Log(s2.Oracles)
// }
