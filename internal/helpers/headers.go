package helpers

import (
	"strings"
)

func ParseExpoExtraParams(header string) map[string]string {
	params := make(map[string]string)
	pairs := strings.Split(header, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], `"`)
		params[key] = value
	}
	return params
}
