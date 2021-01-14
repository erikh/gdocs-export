package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/erikh/gdocs-export/pkg/cli"
	"github.com/erikh/gdocs-export/pkg/converters"
	"google.golang.org/api/docs/v1"
)

func main() {
	var doc docs.Document

	if err := json.NewDecoder(os.Stdin).Decode(&doc); err != nil {
		cli.ErrExit("Could not decode input as JSON: %v", err)
	}

	res, err := converters.Markdown(&doc)
	if err != nil {
		cli.ErrExit("Trouble generating markdown: %v", err)
	}

	fmt.Println(res)
}
