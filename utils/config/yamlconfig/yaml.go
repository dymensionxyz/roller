package yamlconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

func UpdateNestedYAML(node *yaml.Node, path []string, value interface{}) error {
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return fmt.Errorf("empty document node")
		}
		return UpdateNestedYAML(node.Content[0], path, value)
	}

	if len(path) == 0 {
		return setNodeValue(node, value)
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == path[0] {
			return UpdateNestedYAML(node.Content[i+1], path[1:], value)
		}
	}

	// If the path doesn't exist, create it
	newNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: path[0],
		Tag:   "!!str",
	}
	valueNode := &yaml.Node{Kind: yaml.ScalarNode}
	node.Content = append(node.Content, newNode, valueNode)
	return UpdateNestedYAML(valueNode, path[1:], value)
}

func setNodeValue(node *yaml.Node, value interface{}) error {
	switch v := value.(type) {
	case string:
		node.Value = v
		node.Tag = "!!str"
	case int:
		node.Value = strconv.Itoa(v)
		node.Tag = "!!int"
	case float64:
		node.Value = strconv.FormatFloat(v, 'f', -1, 64)
		node.Tag = "!!float"
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}

func CreateYamlNodeFromFile(home, filename string) (*yaml.Node, error) {
	eibcConfigPath := filepath.Join(home, filename)
	data, err := os.ReadFile(eibcConfigPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil, err
	}

	// Parse the YAML
	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		return nil, err
	}

	return &node, nil
}

func PrintYAMLStructure(node *yaml.Node, indent string) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, n := range node.Content {
			PrintYAMLStructure(n, indent)
		}
	case yaml.MappingNode:
		fmt.Printf("%sMapping:\n", indent)
		for i := 0; i < len(node.Content); i += 2 {
			fmt.Printf("%s  %s:\n", indent, node.Content[i].Value)
			PrintYAMLStructure(node.Content[i+1], indent+"    ")
		}
	case yaml.SequenceNode:
		fmt.Printf("%sSequence:\n", indent)
		for _, n := range node.Content {
			PrintYAMLStructure(n, indent+"  ")
		}
	case yaml.ScalarNode:
		fmt.Printf("%sScalar: %s\n", indent, node.Value)
	}
}
