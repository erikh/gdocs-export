package cli

import (
	"fmt"
	"os"
)

// ErrExit is a formatter to stderr, that exits 1.
func ErrExit(f string, items ...interface{}) {
	fmt.Fprintf(os.Stderr, f+"\n", items...)
	os.Exit(1)
}
