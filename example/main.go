package main

import (
	"fmt"
	"log"

	"github.com/divinerapier/imagededup"
)

func main() {
	dedup, err := imagededup.NewImageDedup(1)
	if err != nil {
		log.Fatalf("failed to create image dedup: %v", err)
	}

	defer dedup.Cleanup()
	defer dedup.Close()
	results, err := dedup.FindDuplicates(imagededup.AlgorithmCNN, "tests/data/datasets")
	if err != nil {
		log.Fatalf("failed to find duplicates: %v", err)
	}

	fmt.Println(results)
}
