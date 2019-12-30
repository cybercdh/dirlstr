# dirlstr

Finds Directory Listing's from a list of URLs by traversing the URL paths, e.g.

```
  https://example.com/foo/bar/baz
  https://example.com/foo/bar/
  https://example.com/foo/
```

## Install

If you have Go installed and configured (i.e. with `$GOPATH/bin` in your `$PATH`):

```
go get -u github.com/cybercdh/dirlstr
```

## Usage

```
dirlstr <domain>
$ cat urls.txt | dirlstr
```

## Thanks
This code was heavily inspired by @tomnomnom. 
In the immortal words of Russ Hanneman....."that guy f&ast;&ast;ks"