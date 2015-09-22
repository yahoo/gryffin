// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bitly/go-nsq"

	"github.com/yahoo/gryffin"
	"github.com/yahoo/gryffin/data"
	"github.com/yahoo/gryffin/fuzzer/arachni"
	"github.com/yahoo/gryffin/fuzzer/sqlmap"
	"github.com/yahoo/gryffin/renderer"
)

var storage = flag.String("storage", "memory", "storag method or the storage url")
var service string
var url string
var wg sync.WaitGroup
var wq chan bool

var t *gryffin.Scan

var logWriter io.Writer

// var method = flag.String("method", "GET", "the HTTP method for the request.")
// var url string
// var body = flag.String("data", "", "the data used in a (POST) request.")

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tgryffin-distributed --storage=[memory,redis-url] [seed,crawl,fuzz-sqlmap,fuzz-arachni] [url] \n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

// handler
type h struct {
	HandleMessage nsq.HandlerFunc
}

func captureCtrlC() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)

	go func() {
		<-sigChan
		fmt.Println("We got Ctrl-C. Stopping.")
		wg.Done()
	}()
}

func newProducer() *nsq.Producer {
	producer, err := nsq.NewProducer("127.0.0.1:4150", nsq.NewConfig())
	if err != nil {
		fmt.Println("Cannot connect to NSQ for producing message", err)
		return nil
	}
	return producer
}

func newConsumer(topic, channel string, handler nsq.HandlerFunc) *nsq.Consumer {
	consumer, _ := nsq.NewConsumer(topic, channel, nsq.NewConfig())
	consumer.AddHandler(handler)
	err := consumer.ConnectToNSQLookupd("127.0.0.1:4161")
	if err != nil {
		fmt.Println("Cannot connect to NSQ for consuming message", err)
		return nil
	}
	return consumer
}

func seed(url string) {
	producer := newProducer()
	defer producer.Stop()

	err := t.Poke(&http.Client{})
	if err != nil {
		fmt.Println("Site is not up. Ignoring.", t.Request.URL)
		return
	}

	err = producer.Publish("seed", t.Json())
	if err != nil {
		fmt.Println("Could not publish", "seed", err)
	}
	fmt.Printf("Seed %s injected.\n", url)

}

func crawl() {

	var producer *nsq.Producer
	var consumer *nsq.Consumer

	handler := nsq.HandlerFunc(func(m *nsq.Message) error {
		scan := gryffin.NewScanFromJson(m.Body, t)

		if delay := scan.RateLimit(); delay != 0 {
			go func() {
				time.Sleep(time.Duration(delay) * time.Second)
				err := producer.Publish("seed", scan.Json())
				if err != nil {
					fmt.Println("Could not publish", "fuzz", err)
				}
			}()
		} else {
			// TODO - phantom JS timeout should be an input argument.
			r := &renderer.PhantomJSRenderer{Timeout: 60}
			wq <- true
			scan.CrawlAsync(r)
			go func() {
				if s := <-r.GetRequestBody(); s != nil {
					// fmt.Println("Got request body", s.Request.URL)
					err := producer.Publish("fuzz", s.Json())
					if err != nil {
						fmt.Println("Could not publish", "fuzz", err)
					}
				}
			}()

			go func() {
				isUnique := false
				for s := range r.GetLinks() {
					// do the evaluation once only.
					isUnique = isUnique || scan.IsUnique()
					if isUnique {
						if ok := s.ApplyLinkRules(); ok {
							err := producer.Publish("seed", s.Json())
							if err != nil {
								fmt.Println("Could not publish", "seed", err)
							}
						}
					}
				}
				<-wq
			}()
		}

		return nil
	})

	producer = newProducer()
	defer producer.Stop()
	consumer = newConsumer("seed", "primary", handler)
	defer consumer.Stop()

	wg.Wait()

}

func fuzzWithSqlmap() {
	var consumer *nsq.Consumer
	handler := nsq.HandlerFunc(func(m *nsq.Message) error {
		wq <- true
		scan := gryffin.NewScanFromJson(m.Body, t)
		f := &sqlmap.Fuzzer{}
		f.Fuzz(scan)
		<-wq
		return nil
	})
	consumer = newConsumer("fuzz", "sqlmap", handler)
	defer consumer.Stop()
	wg.Wait()
}

func fuzzWithArachni() {
	var consumer *nsq.Consumer
	handler := nsq.HandlerFunc(func(m *nsq.Message) error {
		wq <- true
		scan := gryffin.NewScanFromJson(m.Body, t)
		f := &arachni.Fuzzer{}
		f.Fuzz(scan)
		<-wq
		return nil
	})
	consumer = newConsumer("fuzz", "arachni", handler)
	defer consumer.Stop()
	wg.Wait()
}

func main() {

	flag.Usage = usage
	flag.Parse()

	switch flag.NArg() {
	case 1:
		// gryffin-distributed crawl
		service = flag.Arg(0)
	case 2:
		// gryffin-distributed seed "http://..."
		service = flag.Arg(0)
		if service == "seed" {
			url = flag.Arg(1)
		} else {
			usage()
			return
		}
	default:
		usage()
		return
	}

	// TCP port listening messages.
	tcpout, err := net.Dial("tcp", "localhost:5000")
	if err != nil {
		// fmt.Println("Cannot establish tcp connection to log listener.")
		logWriter = os.Stdout
	} else {
		logWriter = io.MultiWriter(os.Stdout, tcpout)
	}

	// we use a buffered channel to block when max concurrency is reach.
	maxconcurrency := 5
	wq = make(chan bool, maxconcurrency)

	t = gryffin.NewScan("GET", url, "", data.NewMemoryStore(), logWriter)

	// seed is unique case that we exit the program immediately
	if service == "seed" {
		seed(url)
		return
	}

	captureCtrlC()

	switch service {

	case "crawl":
		crawl()

	case "fuzz-sqlmap":
		fuzzWithSqlmap()
	case "fuzz-arachni":
		fuzzWithArachni()

	default:
		fmt.Println("Unrecognizated service:", service)
		usage()
	}

}
