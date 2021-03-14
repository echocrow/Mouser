// Package config provides configuration parsing for mouser.
package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func newYAMLConfigError(node *yaml.Node, format string, a ...interface{}) error {
	format = "config format error, line %d column %d: " + format
	n := []interface{}{node.Line, node.Column}
	return fmt.Errorf(format, append(n, a...)...)
}

func findChildNode(node *yaml.Node, key string) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, newYAMLConfigError(node, "action item must be a map")
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind == yaml.ScalarNode {
			if k := keyNode.Value; k == key {
				valueNode := node.Content[i+1]
				return valueNode, nil
			}
		}
	}
	return nil, nil
}

// KeyAlias is a key alias used in key mapping.
type KeyAlias string

type mappingKey struct {
	Key string
}

// MappingKey describes a key mapping to a key alias.
type MappingKey mappingKey

// Mapping maps key aliases to keys.
type Mapping map[KeyAlias]MappingKey

// UnmarshalYAML decodes a MappingKey YAML node.
func (mk *MappingKey) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		if err := node.Decode(&mk.Key); err != nil {
			return err
		}
	} else if node.Kind == yaml.MappingNode {
		var m mappingKey
		if err := node.Decode(&m); err != nil {
			return err
		}
		*mk = MappingKey(m)
	} else {
		return newYAMLConfigError(node, "mapping key must be a string or dictionary")
	}
	return nil
}

// Config contains a mouser runtime configuration.
type Config struct {
	Mappings Mapping
	Gestures map[KeyAlias]GestureActions
	Actions  map[string]ActionRef
	Settings Settings
}

// ParseYAML decodes a YAML document into a mouser config.
func ParseYAML(contents []byte) (Config, error) {
	config := Config{
		Settings: DefaultSettings,
	}
	err := yaml.Unmarshal(contents, &config)
	return config, err
}
