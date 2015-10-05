// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package data provides an interface for common data store operations.
package data

// Store is an interface that capture all methods supported for a data store.
type Store interface {
	Get(key string) (value interface{}, ok bool)
	Set(key string, value interface{}) bool
	IncrBy(key string, delta int64) (newVal int64)
	Publish(key string, value interface{})
}
