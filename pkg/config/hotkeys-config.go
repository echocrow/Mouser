package config

import (
	"strings"

	"gopkg.in/yaml.v3"
)

const gestureSeriesStrSep = '.'

func decodeGestureSeriesStr(str string) GestureSeries {
	return strings.Split(str, string(gestureSeriesStrSep))
}

// GestureSeries holds a series of gesture names.
type GestureSeries []string

// UnmarshalYAML decodes a GestureSeries YAML node.
func (gs *GestureSeries) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var gsStr string
		if err := node.Decode(&gsStr); err != nil {
			return err
		}
		gsSeq := decodeGestureSeriesStr(gsStr)
		*gs = gsSeq
	case yaml.SequenceNode:
		var gsSeq []string
		if err := node.Decode(&gsSeq); err != nil {
			return err
		}
		*gs = gsSeq
	default:
		return newYAMLConfigError(node, "gesture sequence must be a string or list")
	}
	if len(*gs) == 0 {
		return newYAMLConfigError(node, "gesture sequence must be not be empty")
	}
	for _, g := range *gs {
		if g == "" {
			return newYAMLConfigError(node, "gesture must be not be empty")
		}
	}
	return nil
}

// GestureAction describes an action associated with a gesture pattern.
type GestureAction struct {
	Gesture GestureSeries
	Exact   bool
	Action  ActionRef
}

// GestureActions holds a slide of GestureAction
type GestureActions []GestureAction

// UnmarshalYAML decodes a GestureActions YAML node.
func (gas *GestureActions) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.SequenceNode:
		var gasSlice []GestureAction
		if err := node.Decode(&gasSlice); err != nil {
			return err
		}
		*gas = gasSlice
	case yaml.MappingNode:
		*gas = make(GestureActions, len(node.Content)/2)
		for i := range *gas {
			k := i * 2
			v := k + 1

			var gs GestureSeries
			keyNode := node.Content[k]
			if err := keyNode.Decode(&gs); err != nil {
				return err
			}

			var a ActionRef
			valueNode := node.Content[v]
			if err := valueNode.Decode(&a); err != nil {
				return err
			}

			(*gas)[i] = GestureAction{Gesture: gs, Action: a}
		}
	default:
		return newYAMLConfigError(node, "hotkey must be a dictionary or list")
	}
	return nil
}
