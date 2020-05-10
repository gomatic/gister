a commandline gister in golang
---
[![GoDoc](https://godoc.org/github.com/gomatic/gister?status.svg)](https://godoc.org/github.com/gomatic/gister)
![](https://img.shields.io/github/issues/gomatic/gister.svg)

> This is a port of [gist](https://github.com/defunkt/gist) in Go.
> Forked from [viyatb/gister](https://github.com/viyatb/gister).

## Settings

1. [Create a personal access token](https://github.com/settings/tokens/new)
1. Set the `GITHUB_TOKEN` environment variable to the value `username:token`
   or write `username:token` to `~/.gist` file.

## Usage

    $ gist -h
    usage: gist [options] file...
      -anonymous
            Set to true for anonymous gist user
      -config string
            Config file. (default "/Users/rnix/.gist")
      -description string
            Description for gist.
      -public
            Set to true for public gist.
      -update string
            Id of existing gist to update.

## LICENSE

[MIT](LICENSE.md)
