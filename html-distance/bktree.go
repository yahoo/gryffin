// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package html-distance is a go library for computing the proximity of the HTML pages. The implementation similiarity fingerprint is Charikar's simhash.
//
// Distance is the hamming distance of the fingerprints. Since fingerprint is of size 64 (inherited from hash/fnv), Similiarity is defined as 1 - d / 64.
//
// In normal scenario, similarity > 95% (i.e. d>3) could be considered as duplicated html pages.
package distance

import (
	"github.com/mfonda/simhash"
)

// Oracle answers the query if a fingerprint has been seen.
type Oracle struct {
	fingerprint uint64      // node value.
	nodes       [65]*Oracle // leaf nodes
}

// NewOracle return an oracle that could tell if the fingerprint has been seen or not.
func NewOracle() *Oracle {
	return newNode(0)
}

func newNode(f uint64) *Oracle {
	return &Oracle{fingerprint: f}
}

// Distance return the similarity distance between two fingerprint.
func Distance(a, b uint64) uint8 {
	return simhash.Compare(a, b)
}

// See asks the oracle to see the fingerprint.
func (n *Oracle) See(f uint64) *Oracle {
	d := Distance(n.fingerprint, f)

	if d == 0 {
		// current node with same fingerprint.
		return n
	}

	// the target node is already set,
	if c := n.nodes[d]; c != nil {
		return c.See(f)
	}

	n.nodes[d] = newNode(f)
	return n.nodes[d]
}

// Seen asks the oracle if anything closed to the fingerprint in a range (r) is seen before.
func (n *Oracle) Seen(f uint64, r uint8) bool {
	d := Distance(n.fingerprint, f)
	if d < r {
		return true
	}

	// TODO - should search from d, d-1, d+1, ... until d-r and d+r, for best performance
	for k := d - r; k <= d+r; k++ {
		if k > 64 {
			break
		}
		if c := n.nodes[k]; c != nil {
			if c.Seen(f, r) == true {
				return true
			}
		}
	}
	return false
}
