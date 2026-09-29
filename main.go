/*

dirlstr
- given a list of urls from stdin, dirlstr will traverse the url paths and look for directory listing.
- where directory listing is found, results are output to the console.
- also checks for an open S3 bucket.

e.g.
$ cat urls.txt | dirlstr

options:

 -c int = Concurrency (default 20; 50 is quick)
 -v = Verbose (for added info)

 written by @cybercdh
 heavily inspired by @tomnomnom. In the immortal words of Russ Hanneman....."that guy f**ks"

*/

package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// maxBodyBytes caps how much of a response is read. A listing marker sits
// near the top of the page, so 1 MiB is plenty and a hostile or huge body
// cannot exhaust memory across the worker pool.
const maxBodyBytes = 1 << 20

// listingMarkers are lower-cased fragments that identify an autoindex page or
// an open storage bucket. Kept specific on purpose: the old bare "Index of"
// check matched any page that happened to contain those words.
var listingMarkers = [][]byte{
	[]byte("<title>index of"),          // Apache, nginx, lighttpd, Caddy
	[]byte("<h1>index of"),             // Apache without a title
	[]byte("index of /"),               // reskinned autoindex pages
	[]byte("directory listing for"),    // Python http.server, Twisted
	[]byte("[to parent directory]"),    // IIS
	[]byte("<listbucketresult xmlns="), // S3 and GCS
	[]byte("<enumerationresults"),      // Azure blob containers
}

func main() {
	var concurrency int
	flag.IntVar(&concurrency, "c", 20, "set the concurrency level")

	var to int
	flag.IntVar(&to, "t", 10000, "timeout (milliseconds)")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Get more info on URL attempts")

	flag.Parse()

	if concurrency < 1 {
		fmt.Fprintf(os.Stderr, "[!]\t-c must be at least 1 (got %d)\n", concurrency)
		os.Exit(2)
	}

	timeout := time.Duration(to) * time.Millisecond

	tr := &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: time.Second,
		}).DialContext,
	}

	// Redirects are deliberately not followed: a 301 from /dir to /dir/ or a
	// bounce to a login page must not be mistaken for the page we asked for.
	// The trailing slash variant is requested explicitly instead (see below).
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}

	urls := make(chan string)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urls {
				if isDirectoryListing(client, u) {
					if verbose {
						fmt.Printf("[*]\tDirectory Listing Found at %s\n", u)
					} else {
						fmt.Println(u)
					}
				}
			}
		}()
	}

	var input io.Reader = os.Stdin
	if arg := flag.Arg(0); arg != "" {
		input = strings.NewReader(arg)
	}

	// Diagnostics go to stderr so stdout stays a clean list of hits when piped.
	seen := make(map[string]bool)
	enqueue := func(u string) {
		if seen[u] {
			if verbose {
				fmt.Fprintf(os.Stderr, "[-]\tAlready seen %s\n", u)
			}
			return
		}
		seen[u] = true
		if verbose {
			fmt.Fprintf(os.Stderr, "[+]\tAttempting: %s\n", u)
		}
		urls <- u
	}

	sc := bufio.NewScanner(input)
	for sc.Scan() {
		for _, u := range expandPaths(sc.Text()) {
			enqueue(u)
		}
	}

	close(urls)

	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[!]\tfailed to read input: %s\n", err)
	}

	wg.Wait()
}

// expandPaths turns one input line into every URL to probe: the URL itself,
// each parent path up to the host root, and the trailing slash form of each
// (servers answer a 301 for /dir, which we do not follow, and the listing for
// /dir/). Blank lines, comments, and unparsable URLs yield nothing. A bare
// host/path without a scheme gets http://.
func expandPaths(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if !strings.Contains(line, "://") {
		line = "http://" + line
	}
	u, err := url.Parse(line)
	if err != nil || u.Host == "" {
		return nil
	}
	base := u.Scheme + "://" + u.Host
	// EscapedPath keeps percent-encoding intact so the request goes out as it
	// was given rather than with a decoded, possibly ambiguous, path.
	parts := strings.Split(u.EscapedPath(), "/")

	var out []string
	add := func(s string) {
		for _, existing := range out {
			if existing == s {
				return
			}
		}
		out = append(out, s)
	}
	for i := 0; i < len(parts); i++ {
		p := strings.Join(parts[:len(parts)-i], "/")
		add(base + p)
		if !strings.HasSuffix(p, "/") {
			add(base + p + "/")
		}
	}
	return out
}

func isDirectoryListing(client *http.Client, target string) bool {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return false
	}
	// set custom UA coz I'm 1337
	req.Header.Set("User-Agent", "dirlstr/1.0")
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// A listing is a 200. Error pages and redirect bodies that mention
	// "Index of" are not hits.
	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return false
	}
	return looksLikeListing(body)
}

// looksLikeListing reports whether body contains a known directory listing or
// bucket listing marker, case insensitively.
func looksLikeListing(body []byte) bool {
	lowered := bytes.ToLower(body)
	for _, m := range listingMarkers {
		if bytes.Contains(lowered, m) {
			return true
		}
	}
	return false
}
