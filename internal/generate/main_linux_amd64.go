package main

import (
	"github.com/divinerapier/go-embed-python/pip"
)

func main() {
	err := pip.CreateEmbeddedPipPackages(
		"requirements.txt",
		"linux",
		"amd64",
		[]string{"manylinux_2_17_x86_64", "manylinux_2_28_x86_64", "manylinux2014_x86_64"},
		"./data/",
	)
	if err != nil {
		panic(err)
	}
}
