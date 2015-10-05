// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package distance

import (
	"bytes"
	"io"

	"github.com/mfonda/simhash"
	"golang.org/x/net/html"
)

// Fingerprint generates the fingerprint of an HTML from the io.Reader r and a shingle factor.
// Shingle refers to the level of shuffling.
// E.g. with shingle factor =2, input "a", "b", "c" will be converted to "a b", "b c"
func Fingerprint(r io.Reader, shingle int) uint64 {
	if shingle < 1 {
		shingle = 1
	}
	// collect the features via this cf channel.
	cf := make(chan string, 1000)
	cs := make(chan uint64, 1000)
	v := simhash.Vector{}

	// Tokenize and then Generate Features. .
	go func() {
		defer close(cf)
		z := html.NewTokenizer(r)
		// TODO - export the max token count as an function argument.
		count := 0
		for tt := z.Next(); count < 5000 && tt != html.ErrorToken; tt = z.Next() {
			t := z.Token()
			count++
			genFeatures(&t, cf)
		}

	}()

	// Collect the features.
	go func() {
		defer close(cs)
		a := make([][]byte, shingle)
		for f := <-cf; f != ""; f = <-cf {
			// shingle: generate the k-gram token as a single feature.
			a = append(a[1:], []byte(f))
			// fmt.Printf("%#v\n", a)
			// fmt.Printf("%s\n", bytes.Join(a, []byte(" ")))
			cs <- simhash.NewFeature(bytes.Join(a, []byte(" "))).Sum()
			// cs <- simhash.NewFeature([]byte(f)).Sum()
		}
	}()

	// from the checksum (of feature), append to vector.
	for s := <-cs; s != 0; s = <-cs {
		for i := uint8(0); i < 64; i++ {
			bit := ((s >> i) & 1)
			if bit == 1 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}

	return simhash.Fingerprint(v)

}

func genFeatures(t *html.Token, cf chan<- string) {

	s := ""

	switch t.Type {
	case html.StartTagToken:
		s = "A:" + t.DataAtom.String()
	case html.EndTagToken:
		s = "B:" + t.DataAtom.String()
	case html.SelfClosingTagToken:
		s = "C:" + t.DataAtom.String()
	case html.DoctypeToken:
		s = "D:" + t.DataAtom.String()
	case html.CommentToken:
		s = "E:" + t.DataAtom.String()
	case html.TextToken:
		s = "F:" + t.DataAtom.String()
	case html.ErrorToken:
		s = "Z:" + t.DataAtom.String()
	}
	// fmt.Println(s)
	cf <- s

	for _, attr := range t.Attr {
		switch attr.Key {
		case "class":
			s = "G:" + t.DataAtom.String() + ":" + attr.Key + ":" + attr.Val
		// case "id":
		// 	s = "G:" + t.DataAtom.String() + ":" + attr.Key + ":" + attr.Val
		case "name":
			s = "G:" + t.DataAtom.String() + ":" + attr.Key + ":" + attr.Val
		case "rel":
			s = "G:" + t.DataAtom.String() + ":" + attr.Key + ":" + attr.Val
		default:
			s = "G:" + t.DataAtom.String() + ":" + attr.Key
		}
		// fmt.Println(s)
		cf <- s
	}

	// fmt.Println(s)

}
