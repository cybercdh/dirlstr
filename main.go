/*

dirlstr
- given a list of urls from stdin, dirlstr will traverse the url paths and look for directory listing.
- where directory listing is found, results are output to the console.

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
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {

	// concurrency flag
	var concurrency int
	flag.IntVar(&concurrency, "c", 20, "set the concurrency level")

	// timeout flag
	var to int
	flag.IntVar(&to, "t", 10000, "timeout (milliseconds)")

	// verbose flag
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Get more info on URL attempts")

	flag.Parse()

	// make an actual time.Duration out of the timeout
	timeout := time.Duration(to * 1000000)

	var tr = &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: time.Second,
		}).DialContext,
	}

	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client := &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       timeout,
	}

	// make a urls channel
	urls := make(chan string)

	// spin up a bunch of workers
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			for url := range urls {

				// if Directory Listing is found, print the URL
				if isDirectoryListing(client, url) {
					fmt.Printf("%s\n",url)
					continue
				}

			}
			wg.Done()
		}()
	}

	var input_urls io.Reader
	input_urls = os.Stdin

	arg_url := flag.Arg(0)
	if arg_url != "" {
		input_urls = strings.NewReader(arg_url)
	}

	// sc := bufio.NewScanner(os.Stdin)
	sc := bufio.NewScanner(input_urls)

	// keep track of urls we've seen
	seen := make(map[string]bool)
	
	for sc.Scan() {

		// parse each url
		_url := url.QueryEscape(sc.Text())

		u,err := url.Parse(_url)
		if err != nil {
			log.Fatal(err)
		}

		// split the paths from the parsed url
		paths := strings.Split(u.Path, "/")

		// iterate over the paths slice to traverse and send to urls channel
		for i := 0; i < len(paths); i++ {
		    path := paths[:len(paths)-i]
		    tmp_url := fmt.Sprintf(u.Scheme + u.Host + strings.Join(path,"/")) + "/"	

		    // if we've seen the url already, keep moving
		    if _, ok := seen[tmp_url]; ok {
		    	if verbose {
		    		fmt.Printf("Already seen %s\n", tmp_url)	
		    	}
		    	continue
		    }

		    // add to seen
		    seen[tmp_url] = true

		    if verbose{
		    	fmt.Printf("Attempting: %s\n",tmp_url)	
		    }
		    
		    // feed the channel
		    urls <- tmp_url
		}

	}

	// once all urls are sent, close the channel
	close(urls)

	// check there were no errors reading stdin (unlikely)
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
	}

	// wait until all the workers have finished
	wg.Wait()

} 

func isDirectoryListing (client *http.Client, url string) bool {
	// perform the GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	// set custom UA coz I'm 1337
	req.Header.Set("User-Agent", "dirlstr/1.0")
	req.Header.Add("Connection", "close")
	req.Close = true

	resp, err := client.Do(req)
	
	// assuming a response, read the body
	if resp != nil {
		
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
				return false
		}

		bodyString := string(bodyBytes)

		// look for Directory Listing, if found return true
		if (strings.Contains(bodyString, "Index of")) {
			return true
		}

	}

	if err != nil {
		return false
	}

	// default return false
	return false
}
