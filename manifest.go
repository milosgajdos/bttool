package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Data struct {
		Src string `yaml:"src"`
		Dst string `yaml:"dst,omitempty"`
	} `yaml:"data"`
	Trackers    []string `yaml:"trackers,omitempty"`
	DhtNodes    []string `yaml:"dhtnodes,omitempty"`
	PieceLength string   `yaml:"piecelength,omitempty"`
	CreatedBy   string   `yaml:"author,omitempty"`
	Comment     string   `yaml:comment,omitempty`
	Encoding    string   `yaml:encoding,omitempty`
	Private     bool     `yaml:private,omitempty`
}

func Parse(manifestPath string) (*Manifest, error) {
	var m Manifest

	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	metaYml, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(metaYml, &m); err != nil {
		return nil, err
	}

	if m.Data.Src == "" {
		return nil, fmt.Errorf("Missing src: src is a mandatory field.")
	}

	if len(m.Trackers) == 0 && len(m.DhtNodes) == 0 {
		return nil, fmt.Errorf("You must specify at least one Tracker or at least one DhtNode")
	}

	if len(m.Trackers) > 0 && len(m.DhtNodes) > 0 {
		return nil, fmt.Errorf("You can't specify both Trackers and DhtNodes")
	}

	return &m, nil
}
