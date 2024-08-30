package middleware

import "strings"

var (
	wildcardstr = repeat('*', 4028)
)

func is(one []byte, two string) bool {
	return strings.HasPrefix(string(one), two)
}

func repeat(s byte, l int) string {
	builder := strings.Builder{}
	builder.Grow(l)
	for i := 0; i <= l; i++ {
		builder.WriteByte(s)
	}
	return builder.String()
}

func deleteKeys(data map[string]any) {
	for key, value := range data {
		dfs(key, value, data, sensitiveInResponse[key])
	}
}

func dfs(ik string, iv any, data map[string]any, keyExists bool) {
	switch ivt := iv.(type) {
	case []any:
		for _, val := range ivt {
			dfs(ik, val, data, keyExists)
		}
	case map[string]any:
		deleteKeys(ivt)
	default:
		if keyExists {
			delete(data, ik)
		}
	}
}

func maskKeys(data map[string]any, sensitive map[string]bool) {
	for key, value := range data {
		modScan(key, value, data, sensitive[key], sensitive)
	}
}

func modScan(ik string, iv any, data map[string]any, keyExists bool, sensitive map[string]bool) {
	switch ivt := iv.(type) {
	case []any:
		for _, val := range ivt {
			modScan(ik, val, data, keyExists, sensitive)
		}
	case map[string]any:
		maskKeys(ivt, sensitive)
	default:
		if keyExists {
			switch val := ivt.(type) {
			case string, []byte:
				data[ik] = wildcardstr[:len(val.(string))]
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				data[ik] = -1
			}
		}
	}
}
