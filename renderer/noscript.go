// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renderer

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	// "sync"

	"github.com/yahoo/gryffin"
	"golang.org/x/net/html"
)

// allow 100 crawling in the machine (regardless of domains)

type NoScriptRenderer struct {
	BaseRenderer
}

func (r *NoScriptRenderer) Do(s *gryffin.Scan) {
	r.chanResponse = make(chan *gryffin.Scan, 10)
	r.chanLinks = make(chan *gryffin.Scan, 10)

	crawl := func() {

		defer close(r.chanResponse)
		defer close(r.chanLinks)

		client := &http.Client{}

		client.Timeout = time.Duration(3) * time.Second

		if response, err := client.Do(s.Request); err == nil {
			s.Response = response
		} else {
			s.Logm("NoScriptRenderer", fmt.Sprintf("error in building request: %s", err))
			return
		}

		s.ReadResponseBody()

		if s.IsDuplicatedPage() {
			return
		}

		tokenizer := html.NewTokenizer(strings.NewReader(s.ResponseBody))

		r.chanResponse <- s

		for {
			t := tokenizer.Next()

			switch t {

			case html.ErrorToken:
				return

			case html.StartTagToken:
				token := tokenizer.Token()
				if token.DataAtom.String() == "a" {
					for _, attr := range token.Attr {
						if attr.Key == "href" {
							link := s.Spawn()
							// TODO - we drop relative URL as it would drop "#".
							// Yet, how about real relative URLs?
							if req, err := http.NewRequest("GET", attr.Val, nil); err == nil {
								if true {
									// || req.URL.IsAbs() {
									link.MergeRequest(req)
									if link.IsScanAllowed() {
										r.chanLinks <- link
									}
								}
								// else {
								// FIXME: ignore relative URL.
								// }
							} else {
								log.Printf("error in building request: %s", err)
							}
						}
					}
				}
			}
		}

		// parse and find links.

	}

	go crawl()
}
