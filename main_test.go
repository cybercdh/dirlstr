package main

import (
	"reflect"
	"testing"
)

func TestExpandPaths(t *testing.T) {
	got := expandPaths("https://example.com/foo/bar/baz")
	want := []string{
		"https://example.com/foo/bar/baz",
		"https://example.com/foo/bar/baz/",
		"https://example.com/foo/bar",
		"https://example.com/foo/bar/",
		"https://example.com/foo",
		"https://example.com/foo/",
		"https://example.com",
		"https://example.com/",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestExpandPaths_SchemelessAndEncoded(t *testing.T) {
	got := expandPaths("  httpbin.org/a%20b/c?x=1 ")
	want := []string{
		"http://httpbin.org/a%20b/c",
		"http://httpbin.org/a%20b/c/",
		"http://httpbin.org/a%20b",
		"http://httpbin.org/a%20b/",
		"http://httpbin.org",
		"http://httpbin.org/",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestExpandPaths_Rejects(t *testing.T) {
	for _, in := range []string{"", "   ", "# comment", "http://", "://x"} {
		if got := expandPaths(in); got != nil {
			t.Errorf("expandPaths(%q) = %v, want nil", in, got)
		}
	}
}

func TestLooksLikeListing(t *testing.T) {
	yes := []string{
		"<html><head><title>Index of /uploads</title></head>",
		"<h1>Index Of /</h1>",
		"<title>Directory listing for /</title>",
		"<pre>[To Parent Directory]</pre>",
		`<?xml version="1.0"?><ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`,
		`<EnumerationResults ServiceEndpoint="https://x.blob.core.windows.net/">`,
	}
	no := []string{
		"<title>Search results: index of refraction</title>",
		"<html>Not Found</html>",
		"",
	}
	for _, b := range yes {
		if !looksLikeListing([]byte(b)) {
			t.Errorf("expected listing: %q", b)
		}
	}
	for _, b := range no {
		if looksLikeListing([]byte(b)) {
			t.Errorf("unexpected listing: %q", b)
		}
	}
}
