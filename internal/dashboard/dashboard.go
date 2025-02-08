package dashboard

import "expo-open-ota/config"

func IsDashboardEnabled() bool {
	return config.GetEnv("USE_DASHBOARD") == "true"
}
