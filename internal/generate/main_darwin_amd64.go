package main

import (
	"github.com/kluctl/go-embed-python/pip"
)

func main() {
	err := pip.CreateEmbeddedPipPackages(
		"requirements.txt",
		"darwin",
		"amd64",
		[]string{"macosx_11_0_x86_64", "macosx_12_0_x86_64"},
		"./data/",
	)
	if err != nil {
		panic(err)
	}
}
