package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"launchpad.net/gommap"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/mmap_span"
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
	mInfoArgs := flagSet.Args()

	var mInfos []*metainfo.MetaInfo
	mInfoOut := os.Stdout

	if len(mInfoArgs) > 0 {
		for _, m := range mInfoArgs {
			mInfo, err := metainfo.LoadFromFile(m)
			if err != nil {
				StdErr("%s", err)
				return 1
			}
			mInfos = append(mInfos, mInfo)
		}
	} else {
		mInfo, err := metainfo.Load(os.Stdin)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
		mInfos = append(mInfos, mInfo)
	}

	if d.outfile != "" {
		var err error
		mInfoOut, err = os.OpenFile(d.outfile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
	}

	for _, m := range mInfos {
		if d.format == "json" {
			e := json.NewEncoder(mInfoOut)
			if err := e.Encode(m); err != nil {
				StdErr("%s", err)
				return 1
			}
		} else {
			fmt.Fprint(mInfoOut, m)
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
	mInfoOut := os.Stdout

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
		if _, err := os.Stat(path); os.IsNotExist(err) {
			StdErr("%s", err)
			return err
		}
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
		b.AddDhtNodes(m.DhtNodes)
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
		mInfoOut, err = os.OpenFile(e.outfile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
	}

	errs, _ := batch.Start(mInfoOut, runtime.NumCPU())
	err, ok := <-errs
	if !ok || err == nil {
		StdErr("%s", err)
		return 1
	}

	return 0
}

type Validate struct {
	verbose bool
	data    string
}

func (v *Validate) Name() string {
	return "validate"
}

func (v *Validate) DefineFlags(flagSet *flag.FlagSet) {
	flagSet.BoolVar(&(v.verbose), "verbose", false, "Print the validation messages")
	flagSet.StringVar(&(v.data), "data", "./", "Path to data to verify the metainfo file against")
}

func (v *Validate) Help() string {
	return fmt.Sprintf("%s METAINFO-FILE", v.Name())
}

// This is heavily inspired by the awesome anacrolix's tools
// https://github.com/anacrolix/torrent/tree/master/cmd/torrent-verify
func (v *Validate) Run(flagSet *flag.FlagSet) int {
	mInfoArgs := flagSet.Args()
	var m *metainfo.MetaInfo
	var err error

	if len(mInfoArgs) > 0 {
		m, err = metainfo.LoadFromFile(mInfoArgs[0])
		if err != nil {
			StdErr("%s", err)
			return 1
		}
	} else {
		m, err = metainfo.Load(os.Stdin)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
	}

	devZero, err := os.Open("/dev/zero")
	if err != nil {
		log.Print(err)
	}
	defer devZero.Close()

	mMapSpan := &mmap_span.MMapSpan{}
	if len(m.Info.Files) > 0 {
		for _, file := range m.Info.Files {
			filename := filepath.Join(append([]string{v.data, m.Info.Name}, file.Path...)...)
			goMMap := fileToMmap(filename, file.Length, devZero)
			mMapSpan.Append(goMMap)
		}
	} else {
		goMMap := fileToMmap(v.data, m.Info.Length, devZero)
		mMapSpan.Append(goMMap)
	}

	for piece := 0; piece < (len(m.Info.Pieces)+sha1.Size-1)/sha1.Size; piece++ {
		expectedHash := m.Info.Pieces[sha1.Size*piece : sha1.Size*(piece+1)]
		if len(expectedHash) == 0 {
			break
		}
		hash := sha1.New()
		_, err := mMapSpan.WriteSectionTo(hash, int64(piece)*m.Info.PieceLength, m.Info.PieceLength)
		if err != nil {
			StdErr("%s", err)
			return 1
		}
		if v.verbose {
			fmt.Println(piece, bytes.Equal(hash.Sum(nil), expectedHash))
		}
	}

	return 0
}

func fileToMmap(filename string, length int64, devZero *os.File) gommap.MMap {
	osFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	mmapFd := osFile.Fd()
	goMMap, err := gommap.MapRegion(mmapFd, 0, length, gommap.PROT_READ, gommap.MAP_PRIVATE)
	if err != nil {
		log.Fatal(err)
	}
	if int64(len(goMMap)) != length {
		log.Printf("file mmap has wrong size: %#v", filename)
	}
	osFile.Close()

	return goMMap
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
