package version

import (
	"fmt"
	"io"
	"os"
)

// Package is the overall, canonical project import path under which the
// package was built.
var Package = "github.com/ccustine/beastie"

// Version indicates which version of the binary is running. This is set to
// the latest release tag by hand, always suffixed by "+unknown". During
// build, it will be replaced by the actual version. The value here will be
// used if the registry is run after a go get based install.
var Version = "v0.0.1+unknown"

func FprintVersion(w io.Writer) {
	fmt.Fprintln(w, os.Args[0], Package, Version)
}

// PrintVersion outputs the version information, from Fprint, to stdout.
func PrintVersion() {
	FprintVersion(os.Stdout)
}
