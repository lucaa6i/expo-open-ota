package dashboard

import "expo-open-ota/config"

func IsDashboardEnabled() bool {
	return config.GetEnv("USE_DASHBOARD") == "true"
}

func ComputeGetBranchesCacheKey() string {
	return "dashboard:request:getBranches"
}

func ComputeGetRuntimeVersionsCacheKey(branch string) string {
	return "dashboard:request:getRuntimeVersions:" + branch
}

func ComputeGetUpdatesCacheKey(branch string, runtimeVersion string) string {
	return "dashboard:request:getUpdates:" + branch + ":" + runtimeVersion
}

func ComputeGetUpdateDetailsCacheKey(branch string, runtimeVersion string, updateID string) string {
	return "dashboard:request:getUpdateDetails:" + branch + ":" + runtimeVersion + ":" + updateID
}
