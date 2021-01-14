package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	intCLI "github.com/erikh/gdocs-export/pkg/cli"
	"github.com/erikh/gdocs-export/pkg/converters"
	"github.com/erikh/gdocs-export/pkg/downloader"
	"github.com/erikh/gdocs-export/pkg/oauth2"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/docs/v1"
)

func main() {
	app := cli.NewApp()

	app.Authors = []*cli.Author{{Email: "github@hollensbe.org", Name: "Erik Hollensbe"}}
	app.Usage = "Fetch google docs and (optionally) convert them to markup formats"

	app.Commands = []*cli.Command{
		{
			Name:      "fetch",
			Usage:     "Download the document and (optionally) convert it",
			ArgsUsage: "[gdocs url]",
			Flags: []cli.Flag{
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
				&cli.StringFlag{
					Name:    "convert",
					Aliases: []string{"c"},
					Usage:   "Convert to various formats; -c help for more",
				},
			},
			Action: fetch,
		},
		{
			Name:      "convert",
			Usage:     "Convert an already-downloaded document from JSON",
			ArgsUsage: "[format] [filename]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "assets-dir",
					Aliases: []string{"a"},
					Usage:   "Where downloaded assets are kept (must exist already, with a manifest.json present)",
				},
			},
			Action: convert,
		},
	}

	if err := app.Run(os.Args); err != nil {
		intCLI.ErrExit("Error: %v", err)
	}
}

func convertFormatHelp() {
	fmt.Println("Formats supported:")
	fmt.Println("md")
	os.Exit(0)
}

func convert(ctx *cli.Context) error {
	if ctx.Args().Get(0) == "help" {
		convertFormatHelp()
	}

	if ctx.Args().Len() != 2 {
		return errors.New("invalid arguments; see --help")
	}

	f, err := os.Open(ctx.Args().Get(1))
	if err != nil {
		return err
	}
	defer f.Close()

	var doc docs.Document

	if err := json.NewDecoder(f).Decode(&doc); err != nil {
		return err
	}

	manifest := downloader.Manifest{}

	if ctx.String("assets-dir") != "" {
		m, err := os.Open(filepath.Join(ctx.String("assets-dir"), "manifest.json"))
		if err != nil {
			return err
		}

		if err := json.NewDecoder(m).Decode(&manifest); err != nil {
			return err
		}
	}

	switch conv := ctx.Args().Get(0); conv {
	case "md":
		res, err := converters.Markdown(&doc, manifest)
		if err != nil {
			return fmt.Errorf("Unable to produce markdown: %v", err)
		}

		fmt.Println(res)
	default:
		return fmt.Errorf("%q is an invalid format. Try `-c help`", conv)
	}

	return nil
}

func fetch(ctx *cli.Context) error {
	if ctx.String("convert") == "help" {
		convertFormatHelp()
	}

	if ctx.Args().Len() != 1 {
		fmt.Fprintln(os.Stderr, "Please provide a google docs url to this command.")
		os.Exit(1)
	}

	client := oauth2.GetClient()

	srv, err := docs.New(client)
	if err != nil {
		return fmt.Errorf("Unable to retrieve Docs client: %v", err)
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

	fmt.Fprintln(os.Stderr, "Downloading assets (this can take a bit)")

	a, err := downloader.New(client)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if ctx.Bool("download") {
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

	switch conv := ctx.String("convert"); conv {
	case "md":
		res, err := converters.Markdown(doc, a.Manifest())
		if err != nil {
			return fmt.Errorf("Unable to produce markdown: %v", err)
		}

		fmt.Println(res)
	case "":
		content, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("Unable to marshal json: %v", err)
		}

		fmt.Println(string(content))
	default:
		return fmt.Errorf("%q is an invalid format. Try `-c help`", conv)
	}

	return nil
}
