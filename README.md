# go-ttlmap

[![Build Status](https://app.travis-ci.com/dtoubelis/go-ttlmap.svg?branch=master)](https://app.travis-ci.com/github/dtoubelis/go-ttlmap)
[![codecov](https://codecov.io/gh/dtoubelis/go-ttlmap/branch/master/graph/badge.svg)](https://codecov.io/gh/dtoubelis/go-ttlmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/dtoubelis/go-ttlmap)](https://goreportcard.com/report/github.com/dtoubelis/go-ttlmap)
[![GoDoc](https://godoc.org/github.com/dtoubelis/go-ttlmap?status.svg)](https://godoc.org/github.com/dtoubelis/go-ttlmap)
![GitHub](https://img.shields.io/github/license/dtoubelis/go-ttlmap)

Go TTL Map is a concurent map with entries expiring after a specified interval. This package
requires `go1.14` or newer.

## Overview

This implementation of TTL Map creates a separate goroutine for each map entry
that takes care of entry expiry.

`PutXXX()` methods parent context as one of the parameters and entries are
safely removed from the map when associated context is canceled.

This design can potentially create a race conditions but measures were taken to
address this issue. In particular, random delay in 0-100,000us range is added
to every TTL to reduce probability of race condition when large number of
entries is added to the map in a rapid succession.

A similar condition may occur on context cancellation but impact of it is
rather negligeable.


## ToDo

- Provide code examples
- Improve documentation
- Develop test for concurency and race conditions

## License

See [LICENSE](LICENSE).