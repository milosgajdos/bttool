package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
)

const MinPieceLength int64 = 32 * 1024

func StdErr(format string, a ...interface{}) {
	out := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, strings.TrimSuffix(out, "\n"))
}

func StdOut(format string, a ...interface{}) {
	out := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stdout, strings.TrimSuffix(out, "\n"))
}

func PieceLength(plHuman string) (int64, error) {
	plUint, err := humanize.ParseBytes(plHuman)
	if err != nil {
		return 0, err
	}

	plInt := int64(plUint)
	if plInt == 0 {
		return 0, fmt.Errorf("Unable to convert piece length size")
	}

	// must be power of 2
	if ((plInt & (plInt - 1)) != 0) || (plInt%MinPieceLength != 0) {
		return 0, fmt.Errorf("Supplied PieceLength: %d is not power of 2", plInt)
	}

	return plInt, nil
}
