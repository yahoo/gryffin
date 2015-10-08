// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlmap

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/yahoo/gryffin"
)

type Fuzzer struct{}

func (s *Fuzzer) Fuzz(g *gryffin.Scan) (count int, err error) {

	var cookies []string

	// for _, c := range g.CookieJar.Cookies(g.Request.URL) {
	for _, c := range g.Cookies {
		cookies = append(cookies, c.String())
	}

	args := []string{
		"--batch",
		"--timeout=2",
		"--retries=3",
		"--crawl=0",
		"--disable-coloring",
		"-o",
		"--text-only",
		// "--threads=4",
		"-v", "0",
		"--level=1",
		"--risk=1",
		"--smart",
		"--fresh-queries",
		"--purge-output",
		"--os=Linux",
		"--dbms=MySQL",
		"--delay=0.1",
		"--time-sec=1",
	}

	// TODO: Post method
	// if g.RequestBody != "" {
	// args = append(args, fmt.Sprintf("--data=..."
	// }

	// only for integer based injection.
	var testable []string
	for k, vs := range g.Request.URL.Query() {
		for _, v := range vs {
			_, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				// query param value is an integer
				testable = append(testable, k)
			}
		}
	}
	if len(testable) > 0 {
		args = append(args, "-p", strings.Join(testable, ","))
	}

	// Cookie
	if len(cookies) > 0 {
		fmt.Println(cookies)
		args = append(args, "--cookie", strings.Join(cookies, ";"))
	}

	args = append(args, "-u", g.Request.URL.String())

	cmd := exec.Command("sqlmap", args...)

	g.Logm("SQLMap.Scan", fmt.Sprintf("Run as %s", cmd.Args))

	output, err := cmd.Output()

	if err != nil {
		return
	}

	count = s.extract(g, string(output))

	g.Logm("SQLMap.Scan", fmt.Sprintf("SQLMap return %t", cmd.ProcessState.Success()))
	return

}

func (s *Fuzzer) extract(g *gryffin.Scan, output string) (count int) {

	for _, l := range strings.Split(output, "\n") {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "Payload: "):
			g.Logm("SQLMap.Findings", l)
			count++
		}
	}

	return
}
