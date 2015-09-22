// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package data

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

// MemoryStore is an implementation for memory based data store.
type RedisStore struct {
	c redis.Conn
}

// Set stores the key value pair.
func (s *RedisStore) Set(key string, value interface{}) bool {
	_, err := s.c.Do("SET", key, value)
	if err != nil {
		fmt.Println("RedisStore Set Error", err)
		return false
	}
	return true
}

// Get retrieve the value pointed by the key.
func (s *RedisStore) Get(key string) (value interface{}, ok bool) {
	reply, err := s.c.Do("GET", key)
	if err != nil {
		return nil, false
	}
	if v, err := redis.Int64(reply, err); err == nil {
		return v, true
	}
	if v, err := redis.String(reply, err); err == nil {
		return v, true
	}
	return nil, false
}

// IncrBy increments the value pointed by key with the delta, and return the new value.
func (s *RedisStore) IncrBy(key string, delta int64) (newVal int64) {
	v, err := redis.Int64(s.c.Do("INCRBY", key, delta))
	if err != nil {
		return 0
	}
	return v

}

func (s *RedisStore) DelPrefix(prefix string) {
	// TODO to be implemented.
}

// NewMemoryStore creates the new store.
func NewRedisStore(rawurl string, options ...redis.DialOption) *RedisStore {
	c, err := redis.DialURL(rawurl, options...)
	if err != nil {
		fmt.Println("NewRedisStore Error", err)
		return nil
	}
	s := RedisStore{c: c}
	return &s
}
