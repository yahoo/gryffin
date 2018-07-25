// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package distance

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

// var input = "<p id=0</p>"
// var input = "<p \t\n iD=\"a&quot;B\"  foo=\"bar\"><EM>te&lt;&amp;;xt</em></p>"
var input = `
<html lang="en" class=" is-copy-enabled">
  <head prefix="og: http://ogp.me/ns# fb: http://ogp.me/ns/fb# object: http://ogp.me/ns/object# article: http://ogp.me/ns/article# profile: http://ogp.me/ns/profile#">
    <meta charset='utf-8'>
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta http-equiv="Content-Language" content="en">
    <meta name="viewport" content="width=1020">
    
    
    <title>net/token_test.go at master · golang/net</title>
    <link rel="search" type="application/opensearchdescription+xml" href="/opensearch.xml" title="GitHub">
    <link rel="fluid-icon" href="https://github.com/fluidicon.png" title="GitHub">
    <link rel="apple-touch-icon" sizes="57x57" href="/apple-touch-icon-114.png">
    <link rel="apple-touch-icon" sizes="114x114" href="/apple-touch-icon-114.png">
    <link rel="apple-touch-icon" sizes="72x72" href="/apple-touch-icon-144.png">
    <link rel="apple-touch-icon" sizes="144x144" href="/apple-touch-icon-144.png">
    <meta property="fb:app_id" content="1401488693436528">
`

func TestCreateFingerprint(t *testing.T) {
	r := strings.NewReader(input)
	f := Fingerprint(r, 2)
	t.Logf("%064b", f)
}

func TestSee(t *testing.T) {

	oracle := NewOracle()

	tests := strings.Split(`<a>b<c>
<a>d<c>
<a><p></a></p>
<a>1<p>2</a>3</p>
<a>1<button>2</a>3</button>
<a>1<b>2</a>3</b>
<a>1<div>2<div>3</a>4</div>5</div>
<table><a>1<p>2</a>3</p>
<b><b><a><p></a>
<b><a><b><p></a>
<a><b><b><p></a>
<p>1<s id="A">2<b id="B">3</p>4</s>5</b>
<table><a>1<td>2</td>3</table>
<table>A<td>B</td>C</table>
<a><svg><tr><input></a>`, "\n")

	for _, test := range tests {
		r := strings.NewReader(test)
		f := Fingerprint(r, 2)
		oracle.See(f)
		t.Logf(" ---- for %064b %s.", f, test)
	}

	for _, test := range tests {
		_ = test
		ntest := "<a>d<c>"
		r := strings.NewReader(ntest)
		f := Fingerprint(r, 2)
		t.Logf("%t for %064b %s.", oracle.Seen(f, 2), f, ntest)
	}

}

func TestSeenWithExternalHTML(t *testing.T) {

	t.Skip("skip htmlsample test ..")
	oracle := NewOracle()

	f1, _ := ioutil.ReadFile("./htmlsamples/flickr001.html")
	f2, _ := ioutil.ReadFile("./htmlsamples/flickr002.html")
	f3, _ := ioutil.ReadFile("./htmlsamples/yahoo001.html")

	{
		r := bytes.NewReader(f1)
		f := Fingerprint(r, 2)
		oracle.See(f)
	}

	{
		r := bytes.NewReader(f2)
		f := Fingerprint(r, 2)
		t.Logf("found? %t", oracle.Seen(f, 2))

	}

	{
		r := bytes.NewReader(f3)
		f := Fingerprint(r, 2)
		t.Logf("found? %t", oracle.Seen(f, 2))

	}

}

func BenchmarkFingerprint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := strings.NewReader(input)
		Fingerprint(r, 2)
	}
}

func BenchmarkFingerprintWithExternalHTML(b *testing.B) {

	b.Skip("Skip external dependent tests.")
	resp, err := http.Get("https://www.yahoo.com/")
	if err != nil {
		b.Fatal(err)
	}
	defer resp.Body.Close()
	input, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(input)
		Fingerprint(r, 2)
	}
}
