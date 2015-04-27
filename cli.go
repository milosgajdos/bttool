package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	commands = append(commands, &Decode{}, &Encode{}, &Validate{}, &Send{})
}

type Decode struct {
	outfile string
	format  string
}

func (d *Decode) Name() string {
	return "decode"
}

func (d *Decode) DefineFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&(d.outfile), "outfile", "", "File to output the decoded information.")
	flagSet.StringVar(&(d.format), "format", "txt", "Which output format to use: txt/json.")
}

func (d *Decode) Help() string {
	return fmt.Sprintf("%s METAINFO-FILE1 METAINFO-FILE2 ...", d.Name())
}

func (d *Decode) Run(flagSet *flag.FlagSet) int {
	minfoArgs := flagSet.Args()

	var minfos []*metainfo.MetaInfo
	minfoOut := os.Stdout

	if len(minfoArgs) > 0 {
		for _, mi := range minfoArgs {
			minfo, err := metainfo.LoadFromFile(mi)
			if err != nil {
				StdErr("%s", err)
				return 1
			}
			minfos = append(minfos, minfo)
		}
	} else {
		minfo, err := metainfo.Load(os.Stdin)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
		minfos = append(minfos, minfo)
	}

	if d.outfile != "" {
		var err error
		minfoOut, err = os.OpenFile(d.outfile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
	}

	for _, mi := range minfos {
		if d.format == "json" {
			e := json.NewEncoder(minfoOut)
			if err := e.Encode(mi); err != nil {
				StdErr("%s", err)
				return 1
			}
		} else {
			fmt.Fprint(minfoOut, mi)
		}
	}

	return 0
}

type Encode struct {
	outfile string
}

func (e *Encode) Name() string {
	return "encode"
}

func (e *Encode) DefineFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&(e.outfile), "outfile", "", "Ecoded torrent metainfo file")
}

func (e *Encode) Help() string {
	return fmt.Sprintf("%s MANIFEST", e.Name())
}

func (e *Encode) Run(flagSet *flag.FlagSet) int {
	manifestArgs := flagSet.Args()

	if len(manifestArgs) < 1 {
		StdErr("%s", "You must specify a manifest file")
		return 1
	}

	manifest := manifestArgs[0]
	minfoOut := os.Stdout

	if _, err := os.Stat(manifest); os.IsNotExist(err) {
		StdErr("%s", err)
		return 1
	}

	m, err := Parse(manifest)
	if err != nil {
		StdErr("Error parsing manifest: %s", err)
		return 1
	}

	b := metainfo.Builder{}
	if err := filepath.Walk(m.Data.Src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		b.AddFile(path)
		return nil
	}); err != nil {
		StdErr("%s", err)
		return 1
	}

	if len(m.Trackers) > 0 {
		b.AddAnnounceGroup(m.Trackers)
	} else {
		// TODO: add nodes support to metainfo package
		fmt.Println("DHT will be used")
		return 1
	}

	if m.Data.Dst != "" {
		b.SetName(m.Data.Dst)
	}

	if m.PieceLength != "" {
		pieceLength, err := PieceLength(m.PieceLength)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
		b.SetPieceLength(pieceLength)
	}

	b.SetPrivate(m.Private)
	b.SetCreationDate(time.Now())

	if m.Comment != "" {
		b.SetComment(m.Comment)
	}

	if m.CreatedBy != "" {
		b.SetCreatedBy(m.CreatedBy)
	}

	if m.Encoding != "" {
		b.SetEncoding(m.Encoding)
	}

	batch, err := b.Submit()
	if err != nil {
		StdErr("%s", err)
		return 1
	}

	if e.outfile != "" {
		var err error
		minfoOut, err = os.OpenFile(e.outfile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
	}

	errs, _ := batch.Start(minfoOut, runtime.NumCPU())
	err, ok := <-errs
	if !ok || err == nil {
		StdErr("%s", err)
		return 1
	}

	return 0
}

type Validate struct {
	verbose bool
}

func (v *Validate) Name() string {
	return "validate"
}

func (v *Validate) DefineFlags(flagSet *flag.FlagSet) {
	flagSet.BoolVar(&(v.verbose), "verbose", false, "Print the validation messages")
}

func (v *Validate) Help() string {
	return fmt.Sprintf("%s METAINFO-FILE", v.Name())
}

func (v *Validate) Run(flagSet *flag.FlagSet) int {
	// TODO: implement validate
	return 0
}

type Send struct {
	tracker string
}

func (s *Send) Name() string {
	return "send"
}

func (s *Send) DefineFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&(s.tracker), "apiserver", "", "API server to send the torrent file to")
}

func (s *Send) Help() string {
	return fmt.Sprintf("%s METAINFO-FILE API-SERVER", s.Name())
}

func (s *Send) Run(flagSet *flag.FlagSet) int {
	// TODO: implement send
	return 0
}
