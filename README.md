# Tideland Go Wait

[![GitHub release](https://img.shields.io/github/release/tideland/go-wait.svg)](https://github.com/tideland/go-wait)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/tideland/go-wait/master/LICENSE)
[![Go Module](https://img.shields.io/github/go-mod/go-version/tideland/go-wait)](https://github.com/tideland/go-wait/blob/master/go.mod)
[![GoDoc](https://godoc.org/tideland.dev/go/actor?status.svg)](https://pkg.go.dev/mod/tideland.dev/go/wait?tab=packages)
![Workflow](https://github.com/tideland/go-wait/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tideland/go-wait)](https://goreportcard.com/report/tideland.dev/go/wait)

## Description

**Tideland Go Wait** provides provides a flexible and controlled waiting for wanted
conditions by polling. The function for testing has to be defined, different tickers
for the polling can be generated for

- simple constant intervals,
- a maximum number of constant intervals,
- a constant number of intervals with a deadline,
- c onstant number of intervals with a timeout, and
- jittering intervals.

Own tickers or changing intervals can be implemented too.

I hope you like it. ;)

## Examples

A simple check for an existing file by polling every second for maximal 30 seconds.

```go
// Tick every second for maximal 30 seconds.
ticker := wait.MakeExpiringIntervalTicker(time.Second, 30*time.Second),
contition := func() (bool, error) {
    _, err := os.Stat(myFile)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, err
    }
    // Found file.
    return true, nil
}

wait.Poll(ctx, ticker, condition)
```

## Contributors

- Frank Mueller (https://github.com/themue / https://github.com/tideland / https://tideland.dev)

