// @license
// Copyright (C) 2018  Dinko Korunic
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package cerrgroup provides synchronization, error propagation, and Context
// cancelation for groups of goroutines working on subtasks of a common task.
// Additionally package provides configurable concurrency limit.
// Package has been forked from "x/sync/errgroup" and modified.
package cerrgroup_test

import (
	"errors"
	"fmt"
	"github.com/dkorunic/findlargedir/cerrgroup"
	"net/http"
	"os"
	"runtime"
	"testing"

	"golang.org/x/net/context"
)

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

type Result string
type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(_ context.Context, query string) (Result, error) {
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

func ExampleGroup_justErrors() {
	cg := cerrgroup.New(runtime.NumCPU())
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url // https://golang.org/doc/faq#closures_and_goroutines
		cg.Go(func() error {
			// Fetch the URL.
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
			}
			return err
		})
	}
	// Wait for all HTTP fetches to complete.
	if err := cg.Wait(); err == nil {
		fmt.Println("Successfully fetched all URLs.")
	}
}

func ExampleGroup_parallel() {
	Google := func(ctx context.Context, query string) ([]Result, error) {
		cg, ctx := cerrgroup.WithContext(ctx, runtime.NumCPU())
		searches := []Search{Web, Image, Video}
		results := make([]Result, len(searches))
		for i, search := range searches {
			i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines
			cg.Go(func() error {
				result, err := search(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := cg.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}
	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}
	// Output:
	// web result for "golang"
	// image result for "golang"
	// video result for "golang"
}

func TestZeroGroup(t *testing.T) {
	err1 := errors.New("errgroup_test: 1")
	err2 := errors.New("errgroup_test: 2")
	cases := []struct {
		errs []error
	}{
		{errs: []error{}},
		{errs: []error{nil}},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}
	for _, tc := range cases {
		cg := cerrgroup.New(runtime.NumCPU())
		var firstErr error
		for i, err := range tc.errs {
			err := err
			cg.Go(func() error { return err })
			if firstErr == nil && err != nil {
				firstErr = err
			}
			if gErr := cg.Wait(); gErr != firstErr {
				t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
					"cg.Wait() = %v; want %v",
					cg, tc.errs[:i+1], err, firstErr)
			}
		}
	}
}

func TestWithContext(t *testing.T) {
	errDoom := errors.New("group_test: doomed")
	cases := []struct {
		errs []error
		want error
	}{
		{want: nil},
		{errs: []error{nil}, want: nil},
		{errs: []error{errDoom}, want: errDoom},
		{errs: []error{errDoom, nil}, want: errDoom},
	}
	for _, tc := range cases {
		cg, ctx := cerrgroup.WithContext(context.Background(), runtime.NumCPU())
		for _, err := range tc.errs {
			err := err
			cg.Go(func() error { return err })
		}
		if err := cg.Wait(); err != tc.want {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"cg.Wait() = %v; want %v",
				cg, tc.errs, err, tc.want)
		}
		canceled := false
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}
		if !canceled {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"ctx.Done() was not closed",
				cg, tc.errs)
		}
	}
}
