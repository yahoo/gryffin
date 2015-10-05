// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package data

import (
	// "log"
	// "reflect"
	"strings"
	"sync/atomic"
)

// MemoryStore is an implementation for memory based data store.
type MemoryStore struct {
	heap map[string]interface{}
}

// Set stores the key value pair.
func (m *MemoryStore) Set(key string, value interface{}) bool {
	switch value.(type) {

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		s, _ := convertIntToPtr(value)
		m.heap[key] = s

	default:
		m.heap[key] = value
	}
	return true
}

// Get retrieve the value pointed by the key.
func (m *MemoryStore) Get(key string) (value interface{}, ok bool) {
	value, ok = m.heap[key]
	switch value.(type) {
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		s, ok := convertPtrToInt(value)
		return s, ok
	default:
		return value, ok
	}
	return value, ok
}

// IncrBy increments the value pointed by key with the delta, and return the new value.
func (m *MemoryStore) IncrBy(key string, delta int64) (newVal int64) {
	newVal = atomic.AddInt64(m.heap[key].(*int64), delta)
	return

}

func (m *MemoryStore) DelPrefix(prefix string) {
	for k, _ := range m.heap {
		if strings.HasPrefix(k, prefix) {
			delete(m.heap, k)
		}
	}
}

// Dummy method
func (m *MemoryStore) Publish(k string, d interface{}) {

}

// NewMemoryStore creates the new store.
func NewMemoryStore() *MemoryStore {
	m := MemoryStore{
		heap: make(map[string]interface{}),
	}
	return &m
}

func convertIntToPtr(v interface{}) (s *int64, ok bool) {
	var t int64

	switch v.(type) {

	case int:
		t = int64(v.(int))
	case int8:
		t = int64(v.(int8))
	case int16:
		t = int64(v.(int16))
	case int32:
		t = int64(v.(int32))
	case int64:
		t = int64(v.(int64))
	case uint:
		t = int64(v.(uint))
	case uint8:
		t = int64(v.(uint8))
	case uint16:
		t = int64(v.(uint16))
	case uint32:
		t = int64(v.(uint32))
	case uint64:
		t = int64(v.(uint64))
	}

	return &t, ok
}

func convertPtrToInt(v interface{}) (s int64, ok bool) {

	switch v.(type) {

	case *int:
		return int64(*(v.(*int))), true
	case *int8:
		return int64(*(v.(*int8))), true
	case *int16:
		return int64(*(v.(*int16))), true
	case *int32:
		return int64(*(v.(*int32))), true
	case *int64:
		return int64(*(v.(*int64))), true

	case *uint:
		return int64(*(v.(*uint))), true
	case *uint8:
		return int64(*(v.(*uint8))), true
	case *uint16:
		return int64(*(v.(*uint16))), true
	case *uint32:
		return int64(*(v.(*uint32))), true
	case *uint64:
		return int64(*(v.(*uint64))), true
	}

	return

}
