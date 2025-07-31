package utils

import "strconv"

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GetFloatValue(data map[string]interface{}, key string) (float64, bool) {
	if val, exists := data[key]; exists {
		switch v := val.(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

func GetInt64Value(data map[string]interface{}, key string) (int64, bool) {
	if val, exists := data[key]; exists {
		switch v := val.(type) {
		case int64:
			return v, true
		case int:
			return int64(v), true
		case float64:
			return int64(v), true
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i, true
			}
		}
	}
	return 0, false
}

func GetStringValue(data map[string]interface{}, key string) (string, bool) {
	if val, exists := data[key]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}
