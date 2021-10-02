package config

import (
	"gopkg.in/yaml.v3"
)

// ActionType describes the type of an action.
type ActionType string

// BasicAction is a basic, pre-defined action.
type BasicAction struct {
	Name string `yaml:"action"`
	Args []interface{}
}

// ToggleAction is a toggable action with customizable timings.
type ToggleAction struct {
	Action      ActionRef
	InitDelay   Ms `yaml:"init-delay"`
	RepeatDelay Ms `yaml:"repeat-delay"`
}

// DefaultToggleAction is a ToggleAction with default settings.
var DefaultToggleAction = ToggleAction{
	InitDelay:   DefaultSettings.Toggles.InitDelay,
	RepeatDelay: DefaultSettings.Toggles.RepeatDelay,
}

// AppBranchAction is a foreground-app-specific action.
type AppBranchAction struct {
	Branches map[string]ActionRef
	Fallback ActionRef
}

// RequireAppAction is a running-app-dependent action.
type RequireAppAction struct {
	App      string
	Do       ActionRef
	Fallback ActionRef
}

func getActionNodeType(node *yaml.Node) (string, error) {
	if typeNode, err := findChildNode(node, "type"); err != nil {
		return "", err
	} else if typeNode == nil {
		return "", nil
	} else if typeNode.Kind != yaml.ScalarNode {
		return "", newYAMLConfigError(node, "action type must be a string")
	} else {
		return typeNode.Value, nil
	}
}

// ActionRef is an reference to a particular type of action.
type ActionRef struct {
	A interface{}
}

// UnmarshalYAML decodes an ActionRef YAML node.
func (ref *ActionRef) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var actionName string
		if err := node.Decode(&actionName); err != nil {
			return err
		}
		ref.A = BasicAction{Name: actionName}
	case yaml.MappingNode:
		var actionType string
		var err error
		if actionType, err = getActionNodeType(node); err != nil {
			return err
		}
		switch actionType {
		case "action", "":
			a := BasicAction{}
			err = node.Decode(&a)
			ref.A = a
		case "toggle":
			a := DefaultToggleAction
			err = node.Decode(&a)
			ref.A = a
		case "app-branch":
			a := AppBranchAction{}
			err = node.Decode(&a)
			ref.A = a
		case "require-app":
			a := RequireAppAction{}
			err = node.Decode(&a)
			ref.A = a
		default:
			err = newYAMLConfigError(node, "unknown action type \"%s\"", actionType)
		}
		if err != nil {
			return err
		}
	default:
		return newYAMLConfigError(node, "action must be a string or dictionary")
	}
	return nil
}
