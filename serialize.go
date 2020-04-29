// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"encoding/json"
	"log"
	"net/http"
)

// NewScanFromJson creates a Scan from the passed JSON blob.
func NewScanFromJson(b []byte) *Scan {
	// ensure we got a memory store..
	if memoryStore == nil {
		memoryStore = NewGryffinStore()
	}

	var scan Scan
	json.Unmarshal(b, &scan)
	return &scan
}

// Json serializes Scan as JSON.
func (s *Scan) Json() []byte {
	ss := &SerializableScan{
		s,
		&SerializableRequest{s.Request, ""},
		&SerializableResponse{
			s.Response,
			&SerializableRequest{s.Request, ""},
		},
	}
	b, err := json.Marshal(ss)
	if err != nil {
		log.Printf("Scan.Json: err=%v", err)
		s.Error("Json", err)
	}
	return b

}

// SerializableScan is a Scan extended with serializable
// request and response fields.
type SerializableScan struct {
	*Scan
	Request  *SerializableRequest
	Response *SerializableResponse
}

// SerializableResponse is a Scan extended with serializable
// response field.
type SerializableResponse struct {
	*http.Response
	Request *SerializableRequest
}

// SerializableRequest is a Scan extended with serializable
// request field.
type SerializableRequest struct {
	*http.Request
	Cancel string
}
