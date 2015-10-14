// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gryffin

import (
	"crypto/rand"
	"fmt"
	"io"
)

// GenRandomID generates a random ID.
func GenRandomID() string {
	// UUID generation is trivial per RSC in https://groups.google.com/d/msg/golang-dev/zwB0k2mpshc/l3zS3oxXuNwJ
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%X", buf)
}
