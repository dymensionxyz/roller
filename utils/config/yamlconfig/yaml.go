package yamlconfig

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func UpdateNestedYAML(filename string, updates map[string]interface{}) error {
	// Read YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse YAML
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return err
	}

	// Update values
	for path, value := range updates {
		keys := strings.Split(path, ".")
		err = setNestedValue(yamlData, keys, value)
		if err != nil {
			return fmt.Errorf("error updating %s: %v", path, err)
		}
	}

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}

	// Write updated YAML back to file
	return os.WriteFile(filename, updatedData, 0o644)
}

func setNestedValue(data map[string]interface{}, keys []string, value interface{}) error {
	for i, key := range keys {
		if i == len(keys)-1 {
			data[key] = value
			return nil
		}

		if _, ok := data[key]; !ok {
			data[key] = make(map[string]interface{})
		}

		nestedMap, ok := data[key].(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to set nested map for key: %s", key)
		}

		data = nestedMap
	}
	return nil
}
