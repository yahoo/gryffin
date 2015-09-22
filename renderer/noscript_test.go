// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renderer

import (
	"testing"
)

func TestNoScriptCrawlAsync(t *testing.T) {
	t.Parallel()
	r := &NoScriptRenderer{}
	testCrawlAsync(t, r)
}
