# Export Google Docs to Markdown, HTML, and more

In this repository lives the `gdexport` command which will fetch Google Docs by URL and attempt to convert them to Markdown or just download them in the JSON representation that Google docs provides. It also provides a minimal UI if you'd prefer to go that route.

The formatting is not perfect yet and may not be for some time. This is still a beta work. Check out the [examples directory](example-output) for some output examples.

## Installation

Check out the [releases page](https://github.com/erikh/gdocs-export/releases).

Building requires a [golang](https://golang.org) version 1.14 or greater currently. Many system packages will not work, so install Golang by hand if you need to.

### OAuth2 Credentials

First, create a [credentials.json per this example](https://developers.google.com/docs/api/quickstart/go) (click on
"enable the Google Docs API" and follow the prompts).

If you do not do the `import-credentials` step below, the `fetch` command **will not work**.

```bash
gdexport import-credentials credentials.json
```

The first time you launch the program to fetch a document, you will be prompted to visit something in your browser and insert a code to STDIN.

### Installing the repository and setting it up

```bash
export GOBIN=$HOME/bin
go get -u github.com/erikh/gdocs-export/...
```

## Usage

There are the following sub-commands:

- `import-credentials`: import a `credentials.json` downloaded from the google docs API. Needed to use the `fetch` command.
- `fetch`: Download a document and optionally convert it. `gdexport help fetch` for more information.
- `convert`: Convert a document on disk. `gdexport help convert` for more information.
- `serve`: Boot the UI to do online conversions. Starts on `http://localhost:4000` by default.

## Notes

- Consolas is the font used to make code blocks. Set the font in gdocs to consolas to enable them.
- Image tags are not `![]()`, they are `<img>` in markdown; this is legal and we can use dimensions safer this way.
- The markdown & html sanitizing code is _not_ safe for automated use. Always validate the docs before you publish them.
- Bullets are differently laid out in html specifically with regards to inline paragraph and image content, because of the differences between gdocs and html in this regard.

## Author

Erik Hollensbe <github@hollensbe.org>
