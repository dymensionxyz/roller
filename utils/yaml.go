package utils

import "fmt"

func SetNestedValue(data map[interface{}]interface{}, keyPath []string, value interface{}) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("empty key path")
	}
	if len(keyPath) == 1 {
		if value == nil {
			delete(data, keyPath[0])
		} else {
			data[keyPath[0]] = value
		}
		return nil
	}
	nextMap, ok := data[keyPath[0]].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("failed to get nested map for key: %s", keyPath[0])
	}
	return SetNestedValue(nextMap, keyPath[1:], value)
}

func GetNestedValue(data map[interface{}]interface{}, keyPath []string) (interface{}, error) {
	if len(keyPath) == 0 {
		return nil, fmt.Errorf("empty key path")
	}
	value, ok := data[keyPath[0]]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", keyPath[0])
	}
	if len(keyPath) == 1 {
		return value, nil
	}
	nextMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get nested map for key: %s", keyPath[0])
	}
	return GetNestedValue(nextMap, keyPath[1:])
}
