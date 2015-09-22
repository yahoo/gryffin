// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dummy

import (
	"fmt"
	"os/exec"

	"github.com/yahoo/gryffin"
)

type Fuzzer struct{}

func (s *Fuzzer) Fuzz(g *gryffin.Scan) (count int, err error) {

	cmd := exec.Command("echo", g.Request.URL.Host)
	_, err = cmd.Output()

	g.Logm("Dummy.Scan", fmt.Sprintf("Echo return %t", cmd.ProcessState.Success()))
	return 0, err

}
