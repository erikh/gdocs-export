package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/erikh/gdocs-export/pkg/cli"
	"github.com/erikh/gdocs-export/pkg/downloader"
	"github.com/erikh/gdocs-export/pkg/oauth2"
	"google.golang.org/api/docs/v1"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Please provide a google docs url to this command.")
		os.Exit(1)
	}

	client := oauth2.GetClient()

	srv, err := docs.New(client)
	if err != nil {
		cli.ErrExit("Unable to retrieve Docs client: %v", err)
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		cli.ErrExit("Unable to parse url: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 4 {
		cli.ErrExit("Invalid URL, cannot parse docID properly")
	}

	docID := parts[3]

	fmt.Fprintln(os.Stderr, "Fetching docID", docID)

	doc, err := srv.Documents.Get(docID).Do()
	if err != nil {
		cli.ErrExit("Unable to retrieve data from document: %v", err)
	}

	content, err := doc.MarshalJSON()
	if err != nil {
		cli.ErrExit("Unable to marshal json: %v", err)
	}

	fmt.Println(string(content))

	if os.Getenv("DOWNLOAD") != "" {
		a, err := downloader.New(client)
		if err != nil {
			cli.ErrExit("%v", err)
		}

		if err := a.Download(os.Getenv("DOWNLOAD"), doc); err != nil {
			cli.ErrExit("trouble downloading: %v", err)
		}
	}
}
