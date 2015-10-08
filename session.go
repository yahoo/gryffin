// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/yahoo/gryffin/data"
	"github.com/yahoo/gryffin/html-distance"
)

type GryffinStore struct {
	Oracles map[string]*distance.Oracle
	Hashes  map[string]bool
	Hits    map[string]int
	store   data.Store
	msgIn   <-chan []byte
	msgOut  chan<- []byte
}

type PublishMessage struct {
	F string      // function, i.e. See or Seen
	T string      // type (kind), i.e. oracle or hash
	K string      // key
	V interface{} // value
}

func NewGryffinStore(msgOut chan<- []byte) *GryffinStore {

	msgIn := make(chan []byte)

	store := GryffinStore{
		Oracles: make(map[string]*distance.Oracle),
		Hashes:  make(map[string]bool),
		Hits:    make(map[string]int),
		msgIn:   msgIn,
		msgOut:  msgOut,
	}
	// start a go rountine to read from the channel
	go func() {
		for jsonPayload := range msgIn {
			fmt.Println("Got this", jsonPayload)
		}
	}()

	return &store
}

func (s *GryffinStore) See(prefix string, kind string, v interface{}) {

	if kind == "oracle" {
		s.oracleSee(prefix, v.(uint64), false)
		return
	}
	if kind == "hash" {
		s.hashesSee(prefix, v.(uint64), false)
		return
	}

	// else, error. TODO.
}

func (s *GryffinStore) Seen(prefix string, kind string, v interface{}, r uint8) bool {

	switch kind {
	case "oracle":
		if oracle, ok := s.Oracles[prefix]; ok {
			return oracle.Seen(v.(uint64), r)
		}
	case "hash":
		k := prefix + "/" + strconv.FormatUint(v.(uint64), 10)
		_, ok := s.Hashes[k]
		return ok
	}
	return false
}

func (s *GryffinStore) oracleSee(prefix string, f uint64, localOnly bool) {
	k := prefix
	// Local update
	oracle, ok := s.Oracles[k]
	if !ok {
		s.Oracles[k] = distance.NewOracle()
		oracle = s.Oracles[k]
	}
	oracle.See(f)

	// Remote update
	if !localOnly && s.msgOut != nil {
		go func() {
			jsonPayload, _ := json.Marshal(&PublishMessage{F: "See", T: "oracle", K: k, V: f})
			s.msgOut <- jsonPayload
		}()
	}
}

func (s *GryffinStore) hashesSee(prefix string, h uint64, localOnly bool) {
	k := prefix + "/" + strconv.FormatUint(h, 10)
	s.Hashes[k] = true
	// Remote update
	if !localOnly && s.msgOut != nil {
		go func() {
			jsonPayload, _ := json.Marshal(&PublishMessage{F: "See", T: "hash", K: k, V: h})
			s.msgOut <- jsonPayload
		}()
	}
}

func (s *GryffinStore) Hit(prefix string) bool {
	// prefix is domain.
	ts := time.Now().Truncate(5 * time.Second).Unix()
	k := prefix + "/" + strconv.FormatInt(ts, 10)
	if v, ok := s.Hits[k]; ok {
		if v >= 5 {
			return false
		}
		s.Hits[k]++
		return true
	}
	s.Hits[k] = 1
	return true
}
