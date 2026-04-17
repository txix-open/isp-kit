package cluster

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"github.com/txix-open/isp-kit/json"
)

var (
	// secretFieldSubstrings contains field name substrings that indicate sensitive data.
	secretFieldSubstrings = map[string]bool{
		"password":    true,
		"secret":      true,
		"token":       true,
		"credentials": true,
	}
	// hidingSecretsEvents contains event names that require secret masking in logs.
	hidingSecretsEvents = map[string]bool{
		ConfigSendConfigWhenConnected: true,
		ConfigSendConfigChanged:       true,
		ModuleSendConfigSchema:        true,
	}
)

// HideSecrets masks sensitive data (passwords, secrets, tokens) in JSON configuration
// by replacing their values with "***". Returns the masked JSON or an error.
func HideSecrets(data []byte) ([]byte, error) {
	config := make(map[string]any)
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config for replacement secret")
	}

	flattenConf := bellows.Flatten(config)

	for key := range flattenConf {
		if flattenConf[key] == "" {
			continue
		}
		for tag := range secretFieldSubstrings {
			if strings.Contains(strings.ToLower(key), tag) {
				flattenConf[key] = "***"
			}
		}
	}

	expandConf := bellows.Expand(flattenConf)
	if expandConf == nil {
		expandConf = make(map[string]any)
	}

	config, ok := expandConf.(map[string]any)
	if !ok {
		return nil, errors.WithMessagef(err, "unexpected type from bellows, expected map, got %T", config)
	}

	data, err = json.Marshal(config)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal config for replacement secret")
	}

	return data, nil
}

// RegisterSecretSubstrings adds custom field name substrings to the list of patterns
// that indicate sensitive data.
func RegisterSecretSubstrings(substrings []string) {
	for _, substring := range substrings {
		secretFieldSubstrings[strings.ToLower(substring)] = true
	}
}

// UnregisterSecretSubstrings removes custom field name substrings from the list of
// patterns that indicate sensitive data.
func UnregisterSecretSubstrings(substrings []string) {
	for _, substring := range substrings {
		delete(secretFieldSubstrings, strings.ToLower(substring))
	}
}

// hideSecrets masks sensitive data in logs for events that contain configuration data.
func hideSecrets(event string, data []byte) []byte {
	if hidingSecretsEvents[event] {
		dataToLog, err := HideSecrets(data)
		if err != nil {
			return data
		}
		return dataToLog
	}
	return data
}
