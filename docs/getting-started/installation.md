# Installation

`builder` is a single Go binary. Two ways to get it.

## `go install`

```sh
go install github.com/tricky-bits/builder@latest
```

Puts `builder` on your `$GOPATH/bin` (or `$GOBIN`).

## Build from source

```sh
git clone https://github.com/tricky-bits/builder.git
cd builder
go build -o builder .
```

Produces a `builder` binary in the repository root.

There are no pre-built release binaries yet, one of the two methods above is
required.

Next: [Quick Start](quick-start.md).
