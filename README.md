# Export Google Docs to Markdown, and more

In this repository lives the `gdexport` command which will fetch Google Docs by
URL and attempt to convert them to Markdown or just download them in the JSON
representation that Google docs provides.

The formatting is not perfect yet and may not be for some time. This is still
an early alpha work.

## Installation

Installation requires a [golang](https://golang.org) version 1.14 or greater
currently. Many system packages will not work, so install Golang by hand if you
need to.

### OAuth2 Credentials

First, create a [credentials.json per this example](https://developers.google.com/docs/api/quickstart/go) (click on
"enable the Google Docs API" and follow the prompts).

The first time you launch the program, you will be prompted to visit something
in your browser and insert a code to STDIN. This will save a second file, named
`token.json`. Keep in with `credentials.json`.

### Installing the repository and setting it up

This will become easier with time.

```bash
go get -d github.com/erikh/gdocs-export
mv credentials.json $GOPATH/src/github.com/erikh/gdocs-export
cd $GOPATH/src/github.com/erikh/gdocs-export
go run ./cmd/gdexport --help
```

## Usage

There are the following sub-commands:

- `fetch`: Download a document and optionally convert it. `go run ./cmd/gdexport help fetch` for more information.
- `convert`: Convert a document on disk. `go run ./cmd/gdexport help convert` for more information.

## Author

Erik Hollensbe <github@hollensbe.org>
