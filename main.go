package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/erikh/gdocs-export/pkg/cli"
	"github.com/erikh/gdocs-export/pkg/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Please provide a google docs url to this command.")
		os.Exit(1)
	}

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		cli.ErrExit("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/documents.readonly")
	if err != nil {
		cli.ErrExit("Unable to parse client secret file to config: %v", err)
	}
	client := oauth2.GetClient(config)

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
}
