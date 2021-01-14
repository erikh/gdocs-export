package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	intCLI "github.com/erikh/gdocs-export/pkg/cli"
	"github.com/erikh/gdocs-export/pkg/downloader"
	"github.com/erikh/gdocs-export/pkg/oauth2"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/docs/v1"
)

func main() {
	app := cli.NewApp()

	app.Authors = []*cli.Author{{Email: "github@hollensbe.org", Name: "Erik Hollensbe"}}
	app.Usage = "Fetch google docs and (optionally) convert them to markup formats"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "assets-dir",
			Aliases: []string{"a"},
			Usage:   "Where to put downloaded assets",
			Value:   "./assets",
		},
		&cli.BoolFlag{
			Name:    "download",
			Aliases: []string{"d", "dl"},
			Usage:   "Whether or not to download assets",
		},
	}

	app.Action = action

	if err := app.Run(os.Args); err != nil {
		intCLI.ErrExit("Error: %v", err)
	}
}

func action(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		fmt.Fprintln(os.Stderr, "Please provide a google docs url to this command.")
		os.Exit(1)
	}

	client := oauth2.GetClient()

	srv, err := docs.New(client)
	if err != nil {
		return fmt.Errorf("Unable to retrieve Docs intCLIent: %v", err)
	}

	u, err := url.Parse(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("Unable to parse url: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 4 {
		return fmt.Errorf("Invalid URL, cannot parse docID properly")
	}

	docID := parts[3]

	fmt.Fprintln(os.Stderr, "Fetching docID", docID)

	doc, err := srv.Documents.Get(docID).Do()
	if err != nil {
		return fmt.Errorf("Unable to retrieve data from document: %v", err)
	}

	content, err := doc.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Unable to marshal json: %v", err)
	}

	fmt.Println(string(content))

	if ctx.Bool("download") {
		a, err := downloader.New(client)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		dl := ctx.String("assets-dir")

		if err := a.Download(dl, doc); err != nil {
			return fmt.Errorf("trouble downloading to %q: %v", dl, err)
		}

		manifest, err := a.ManifestJSON()
		if err != nil {
			return fmt.Errorf("Error marshalling manifest: %v", err)
		}

		if err := ioutil.WriteFile(filepath.Join(dl, "manifest.json"), manifest, 0600); err != nil {
			return fmt.Errorf("Error writing manifest to %q: %v", dl, err)
		}
	}

	return nil
}
