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
