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
}

// IncrBy increments the value pointed by key with the delta, and return the new value.
func (m *MemoryStore) IncrBy(key string, delta int64) (newVal int64) {
	newVal = atomic.AddInt64(m.heap[key].(*int64), delta)
	return

}

func (m *MemoryStore) DelPrefix(prefix string) {
	for k := range m.heap {
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

	switch v := v.(type) {

	case int:
		t = int64(v)
	case int8:
		t = int64(v)
	case int16:
		t = int64(v)
	case int32:
		t = int64(v)
	case int64:
		t = v
	case uint:
		t = int64(v)
	case uint8:
		t = int64(v)
	case uint16:
		t = int64(v)
	case uint32:
		t = int64(v)
	case uint64:
		t = int64(v)
	}

	return &t, ok
}

func convertPtrToInt(v interface{}) (s int64, ok bool) {

	switch v := v.(type) {

	case *int:
		return int64(*v), true
	case *int8:
		return int64(*v), true
	case *int16:
		return int64(*v), true
	case *int32:
		return int64(*v), true
	case *int64:
		return *v, true

	case *uint:
		return int64(*v), true
	case *uint8:
		return int64(*v), true
	case *uint16:
		return int64(*v), true
	case *uint32:
		return int64(*v), true
	case *uint64:
		return int64(*v), true
	}

	return

}
