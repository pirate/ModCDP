//go:build ignore

// Go CDP generated-surface entrypoint.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Go CDP codegen is not implemented yet; generated output lives in go/modcdp/client/generated*.go.")
	os.Exit(1)
}
