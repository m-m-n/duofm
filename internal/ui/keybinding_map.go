package ui

import (
	"github.com/sakura/duofm/internal/config"
)

// KeybindingMap maps key strings to actions.
type KeybindingMap struct {
	keyToAction map[string]Action
}

// NewKeybindingMap creates a KeybindingMap from the given configuration.
func NewKeybindingMap(cfg *config.Config) *KeybindingMap {
	km := &KeybindingMap{
		keyToAction: make(map[string]Action),
	}

	if cfg == nil || cfg.Keybindings == nil {
		return km
	}

	for actionName, keys := range cfg.Keybindings {
		action := ActionFromName(actionName)
		if action == ActionNone {
			// Unknown action, skip
			continue
		}

		for _, key := range keys {
			normalized, err := config.NormalizeKey(key)
			if err != nil {
				// Invalid key, skip
				continue
			}
			km.keyToAction[normalized] = action
		}
	}

	return km
}

// DefaultKeybindingMap creates a KeybindingMap with default keybindings.
func DefaultKeybindingMap() *KeybindingMap {
	cfg := &config.Config{
		Keybindings: config.DefaultKeybindings(),
	}
	return NewKeybindingMap(cfg)
}

// GetAction returns the action for the given key, or ActionNone if not found.
func (km *KeybindingMap) GetAction(key string) Action {
	if km == nil || km.keyToAction == nil {
		return ActionNone
	}
	if action, ok := km.keyToAction[key]; ok {
		return action
	}
	return ActionNone
}

// HasKey returns true if the given key is mapped to an action.
func (km *KeybindingMap) HasKey(key string) bool {
	if km == nil || km.keyToAction == nil {
		return false
	}
	_, ok := km.keyToAction[key]
	return ok
}
