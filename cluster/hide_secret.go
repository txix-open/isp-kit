package cluster

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"github.com/txix-open/isp-kit/json"
)

var (
	tagConfigSecrets    = []string{"password", "secret", "token"}
	hidingSecretsEvents = map[string]bool{
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
		for _, tag := range tagConfigSecrets {
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

func RegisterTagsSecrets(customTagConfigSecrets []string) {
	tagConfigSecrets = append(tagConfigSecrets, customTagConfigSecrets...)
}

func UnregisterTagsSecrets(tagsToRemove []string) {
	removeTagsMap := make(map[string]bool)
	for _, tag := range tagsToRemove {
		removeTagsMap[tag] = true
	}

	newTags := []string{}
	for _, tag := range tagConfigSecrets {
		found := removeTagsMap[tag]
		if !found {
			newTags = append(newTags, tag)
		}
	}
	tagConfigSecrets = newTags
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
