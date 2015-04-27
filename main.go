package main

import "os"

var commands []subCmd

func main() {
	runCmd := ParseCmd()
	os.Exit(runCmd())
}
