package interceptor

import "strings"

var (
	wildcardstr = repeat('*', 2024)
)

func DeleteKeys(data map[string]any) {
	deleteKeys(data)
}

func MaskSensitiveKeys(data map[string]any, sensitiveKeys map[string]bool) {
	mask(data, sensitiveKeys)
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
		clean(key, value, data, sensitiveInResponse[key])
	}
}

func clean(ik string, iv any, data map[string]any, keyExists bool) {
	switch ivt := iv.(type) {
	case []any:
		for _, val := range ivt {
			clean(ik, val, data, keyExists)
		}
	case map[string]any:
		deleteKeys(ivt)
	default:
		if keyExists {
			delete(data, ik)
		}
	}
}

func mask(data map[string]any, sensitive map[string]bool) {
	for key, value := range data {
		scan(key, value, data, sensitive[key], sensitive)
	}
}

func scan(ik string, iv any, data map[string]any, keyExists bool, sensitive map[string]bool) {
	switch ivt := iv.(type) {
	case []any:
		for _, val := range ivt {
			scan(ik, val, data, keyExists, sensitive)
		}
	case map[string]any:
		mask(ivt, sensitive)
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
