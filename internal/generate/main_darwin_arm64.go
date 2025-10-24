package main

import (
	"github.com/kluctl/go-embed-python/pip"
)

func main() {
	err := pip.CreateEmbeddedPipPackages(
		"requirements.txt",
		"darwin",
		"arm64",
		[]string{"macosx_11_0_arm64", "macosx_12_0_arm64"},
		"./data/",
	)
	if err != nil {
		panic(err)
	}
}
