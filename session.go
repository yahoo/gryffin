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
	msgOut  chan<- []byte
}

type PublishMessage struct {
	T string
	K string
	V interface{}
}

func NewGryffinStore(msgIn <-chan []byte, msgOut chan<- []byte) *GryffinStore {

	store := GryffinStore{
		Oracles: make(map[string]*distance.Oracle),
		Hashes:  make(map[string]bool),
		Hits:    make(map[string]int),
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

func (s *GryffinStore) See(scan *Scan, v interface{}) {

	switch v.(type) {
	case uint64:
		s.oracleSee(scan, v.(uint64), false)
	case string:
		s.hashesSee(scan, v.(string), false)
	}
}

func (s *GryffinStore) Seen(scan *Scan, v interface{}, r uint8) bool {

	k := scan.Job.ID
	switch v.(type) {
	case uint64:
		f := v.(uint64)
		if oracle, ok := s.Oracles[k]; ok {
			return oracle.Seen(f, r)
		}
	case string:
		h := v.(string)
		_, ok := s.Hashes[k+"/"+h]
		return ok
	}
	return false
}

func (s *GryffinStore) oracleSee(scan *Scan, f uint64, localOnly bool) {
	k := scan.Job.ID
	// Local update
	oracle, ok := s.Oracles[k]
	if !ok {
		s.Oracles[k] = distance.NewOracle()
		oracle = s.Oracles[k]
	}
	oracle.See(f)

	// Remote update
	if !localOnly && s.msgOut != nil {
		jsonPayload, _ := json.Marshal(&PublishMessage{T: "See", K: k, V: f})
		s.msgOut <- jsonPayload
		// s.publisher.Publish("shared-memory", jsonPayload)
	}
}

func (s *GryffinStore) hashesSee(scan *Scan, hash string, localOnly bool) {
	k := scan.Job.ID
	s.Hashes[k+"/"+hash] = true
	// Remote update
	if !localOnly && s.msgOut != nil {
		jsonPayload, _ := json.Marshal(&PublishMessage{T: "See", K: hash, V: true})
		s.msgOut <- jsonPayload
	}
}

func (s *GryffinStore) Hit(domain string) bool {
	ts := time.Now().Truncate(5 * time.Second).Unix()
	k := domain + "/" + strconv.FormatInt(ts, 10)
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
