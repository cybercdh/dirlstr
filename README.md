# dirlstr

Finds Directory Listings or Open S3 Buckets from a list of URLs by traversing the URL paths, e.g.

```
  https://example.com/foo/bar/baz
  https://example.com/foo/bar/
  https://example.com/foo/
```

## Install

If you have Go installed and configured (i.e. with `$GOPATH/bin` in your `$PATH`):

```
go install github.com/cybercdh/dirlstr@latest
```

## Usage

```
$ dirlstr <url>
```
or 
```
$ cat <file> | dirlstr
```

If a URL is found to expose a Directory Listing or an open storage bucket (S3, GCS, or Azure), it will be printed to the console. Every parent path is checked in both `/dir` and `/dir/` form, since most servers redirect the first to the second and redirects are not followed. Verbose output goes to stderr, so stdout stays a clean list of hits when piped.

### Options

```
Usage of dirlstr:
  -c int
    	set the concurrency level (default 20)
  -t int
    	timeout (milliseconds) (default 10000)
  -v	Get more info on URL attempts
```

## Thanks
This code was heavily inspired by [@tomnomnom.](https://github.com/tomnomnom) 
In the immortal words of Russ Hanneman....."that guy f&ast;&ast;ks"
