package cluster

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"github.com/txix-open/isp-kit/json"
)

var (
	secretFieldSubstrings = map[string]struct{}{"password": {}, "secret": {}, "token": {}}
	hidingSecretsEvents   = map[string]bool{
		ConfigSendConfigWhenConnected: true,
		ConfigSendConfigChanged:       true,
		ModuleSendConfigSchema:        true,
	}
)

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

func RegisterSecretSubstrings(substrings []string) {
	for _, substring := range substrings {
		secretFieldSubstrings[strings.ToLower(substring)] = struct{}{}
	}
}

func UnregisterSecrets(substrings []string) {
	for _, substring := range substrings {
		delete(secretFieldSubstrings, strings.ToLower(substring))
	}
}

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
