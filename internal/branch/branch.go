package branch

import (
	"expo-open-ota/internal/helpers"
	"expo-open-ota/internal/services"
)

func UpsertBranch(branch string) error {
	branches, err := services.FetchExpoBranches()
	if err != nil {
		return err
	}
	if !helpers.StringInSlice(branch, branches) {
		return services.CreateBranch(branch)
	}
	return nil
}
