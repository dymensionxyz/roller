package yamlconfig

import (
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

func UpdateNestedYAML(node *yaml.Node, path []string, value string) error {
	if len(path) == 0 {
		node.Value = value
		return nil
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == path[0] {
			return UpdateNestedYAML(node.Content[i+1], path[1:], value)
		}
	}

	return fmt.Errorf("path not found: %v", path[0])
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
