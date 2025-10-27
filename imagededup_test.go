package imagededup_test

import (
	"testing"

	"github.com/divinerapier/imagededup"
	"github.com/stretchr/testify/require"
)

func TestFindDuplicates(t *testing.T) {
	dedup, err := imagededup.NewImageDedup(1)
	if err != nil {
		t.Fatalf("failed to create image dedup: %v", err)
	}

	defer dedup.Cleanup()

	results, err := dedup.FindDuplicates(imagededup.AlgorithmCNN, "tests/data/datasets")
	require.NoError(t, err)
	require.NotEmpty(t, results)
}
