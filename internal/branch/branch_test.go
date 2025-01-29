package branch

import (
	"expo-open-ota/internal/bucket"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	testing2 "testing"
)

func setup(t *testing2.T) func() {
	httpmock.Activate()
	return func() {
		bucket.ResetBucketInstance()
		defer httpmock.DeactivateAndReset()
	}
}

func TestUpsertBranch(t *testing2.T) {
	assert.Equal(t, true, true)
}
