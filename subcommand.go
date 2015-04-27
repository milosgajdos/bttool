// A simple sub command parser based on the flag package
// https://gist.github.com/srid/3949446
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type subCmd interface {
	Name() string
	DefineFlags(*flag.FlagSet)
	Help() string
	Run(*flag.FlagSet) int
}

type subCmdParser struct {
	cmd subCmd
	fs  *flag.FlagSet
}

func ParseCmd() func() int {
	scp := make(map[string]*subCmdParser, len(commands))
	for _, cmd := range commands {
		name := cmd.Name()
		scp[name] = &subCmdParser{cmd, flag.NewFlagSet(name, flag.ExitOnError)}
		cmd.DefineFlags(scp[name].fs)
	}
	oldUsage := flag.Usage
	flag.Usage = func() {
		oldUsage()
		for _, sc := range scp {
			fmt.Fprintf(os.Stderr, "\n%s %s\n", filepath.Base(os.Args[0]), sc.cmd.Help())
			sc.fs.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	return func() int {
		cmdname := flag.Arg(0)
		if sc, ok := scp[cmdname]; ok {
			sc.fs.Parse(flag.Args()[1:])
			return sc.cmd.Run(sc.fs)
		} else {
			fmt.Errorf("\nERROR: %s is not a valid command\n\n", cmdname)
			flag.Usage()
			return 1
		}
	}
}
