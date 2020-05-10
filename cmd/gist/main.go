// This app is intended to be go-port of the defunckt's gist library in Ruby
// Currently, uploading single and multiple files are available.
// You can also create secret gists, and both anonymous and user gists.
//
// Author: Viyat Bhalodia
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomatic/gister/internal/gist"
)

// Defines basic usage when program is run with the help flag
func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: gist [options] file...\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// The main function parses the CLI args. It also checks the files, and
// loads them into an array.
// Then the files are separated into GistFile structs and collectively
// the files are saved in `files` field in the Gist struct.
// A request is then made to the GitHub api - it depends on whether it is
// anonymous gist or not.
// The response recieved is parsed and the Gist URL is printed to STDOUT.
func main() {
	// User agent defines a custom agent (required by GitHub)
	// `token` stores the GITHUB_TOKEN from the env variables
	// GITHUB_TOKEN must be in format of `username:token`
	configFile := os.Getenv("GIST_CONFIG")
	if configFile == "" {
		configFile = filepath.Join(os.Getenv("HOME"), ".gist")
	}

	var (
		anonymous   bool
		public      bool
		description string
		update      string
	)

	flag.BoolVar(&anonymous, "anonymous", false, "Set to true for anonymous gist user")
	flag.StringVar(&update, "update", "", "Id of existing gist to update.")
	flag.BoolVar(&public, "public", false, "Set to true for public gist.")
	flag.StringVar(&description, "description", "", "Description for gist.")
	flag.StringVar(&configFile, "config", configFile, "Config file.")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	if !public && anonymous {
		log.Fatalln("incompatible: private and anonymous")
	}

	token, err := ioutil.ReadFile(configFile)
	if err != nil && !anonymous {
		log.Fatalf("no token. %s: %s", configFile, err)
	}

	gister, err := gist.New(
		gist.Anonymous(anonymous),
		gist.MustFiles(flag.Args()...), // TODO support stdin
		gist.Credentials(strings.TrimSpace(string(token))),
		gist.Description(description),
		gist.Public(public),
		gist.GistId(update),
	)
	if err != nil {
		log.Fatal(err)
	}

	response, err := gister.Send()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.HtmlUrl)
}
