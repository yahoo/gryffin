// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"encoding/json"
	"net/http"
)

func NewScanFromJson(b []byte) *Scan {
	// ensure we got a memory store..
	if memoryStore == nil {
		memoryStore = NewGryffinStore()
	}

	var scan Scan
	json.Unmarshal(b, &scan)
	return &scan
}

// func (s *Scan) Json() []byte {
// 	ss := &SerializableScan{
// 		s,
// 		&SerializableRequest{s.Request, ""},
// 		&SerializableResponse{
// 			s.Response,
// 			&SerializableRequest{s.Request, ""},
// 		},
// 	}
// 	log.Printf("DMDEBUG ss=%#v", ss)
// 	b, err := json.Marshal(ss)
// 	if err != nil {
// 		log.Printf("DMDEBUG error in json.Marshal: %v", err)
// 		s.Error("Json", err)
// 	}
// 	return b

// }

type SerializableScan struct {
	*Scan
	Request  *SerializableRequest
	Response *SerializableResponse
}

type SerializableResponse struct {
	*http.Response
	Request *SerializableRequest
}
type SerializableRequest struct {
	*http.Request
	Cancel string
}
