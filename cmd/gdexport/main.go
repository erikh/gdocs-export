package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	intCLI "github.com/erikh/gdocs-export/pkg/cli"
	"github.com/erikh/gdocs-export/pkg/converters"
	"github.com/erikh/gdocs-export/pkg/downloader"
	"github.com/erikh/gdocs-export/pkg/oauth2"
	"github.com/erikh/gdocs-export/pkg/util"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/docs/v1"
)

func main() {
	app := cli.NewApp()

	app.Authors = []*cli.Author{{Email: "github@hollensbe.org", Name: "Erik Hollensbe"}}
	app.Usage = "Fetch google docs and (optionally) convert them to markup formats"

	app.Commands = []*cli.Command{
		{
			Name:      "import-credentials",
			Usage:     "Import your credentials.json provided to you by google",
			ArgsUsage: "[credentials.json file]",
			Aliases:   []string{"i", "creds"},
			Action:    importCredentials,
		},
		{
			Name:      "fetch",
			Usage:     "Download the document and (optionally) convert it",
			ArgsUsage: "[gdocs url]",
			Aliases:   []string{"f", "download"},
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
			Aliases:   []string{"c", "transform"},
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
	fmt.Println("md, html")
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

	res, err := converters.Convert(ctx.Args().Get(0), &doc, manifest)
	if err != nil {
		return err
	}

	fmt.Println(res)
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

	docID, err := util.ParseDocsURL(ctx.Args().First())
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Fetching docID", docID)

	client := oauth2.GetClient()

	srv, err := docs.New(client)
	if err != nil {
		return fmt.Errorf("Unable to retrieve Docs client: %v", err)
	}

	doc, err := srv.Documents.Get(docID).Do()
	if err != nil {
		return fmt.Errorf("Unable to retrieve data from document: %v", err)
	}

	var manifest downloader.Manifest

	if ctx.Bool("download") {
		dl := ctx.String("assets-dir")
		fmt.Fprintf(os.Stderr, "Downloading assets to %q (this can take a bit)\n", dl)

		var err error
		manifest, err = util.DownloadAssets(client, doc, dl, true)
		if err != nil {
			return err
		}
	}

	if ctx.String("convert") == "" {
		content, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("Unable to marshal json: %v", err)
		}

		fmt.Println(string(content))
	} else {
		res, err := converters.Convert(ctx.String("convert"), doc, manifest)
		if err != nil {
			return err
		}

		fmt.Println(res)
	}

	return nil
}

func importCredentials(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		return errors.New("invalid arguments: please see help")
	}

	f, err := os.Open(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("Cannot open %q: %w", ctx.Args().First(), err)
	}
	defer f.Close()

	return oauth2.ImportCredentials(f)
}
