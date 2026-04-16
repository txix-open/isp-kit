package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"gopkg.in/yaml.v3"
)

// Source defines an interface for configuration providers.
// Implementations should return a flat map of key-value pairs.
// Nested YAML structures are automatically flattened using dot notation.
type Source interface {
	Config() (map[string]string, error)
}

// YamlFileSource is a configuration source that reads from a YAML file.
// It supports nested structures which are flattened into dot-notation keys.
// For example, a YAML key "server.host" will be stored as "server.host".
type YamlFileSource struct {
	file string
}

// NewYamlConfig creates a new YamlFileSource for the specified YAML file.
// The file path can be absolute or relative to the working directory.
func NewYamlConfig(file string) YamlFileSource {
	return YamlFileSource{file: file}
}

// Config loads and parses the YAML file, returning a flat map of configuration values.
// Nested YAML structures are flattened using bellows.Flatten.
// All values are converted to strings.
// Returns an error if the file cannot be opened or parsed as valid YAML.
func (y YamlFileSource) Config() (map[string]string, error) {
	f, err := os.Open(y.file)
	if err != nil {
		return nil, errors.WithMessagef(err, "open %s", y.file)
	}
	defer f.Close()

	fileProps := make(map[string]any)
	err = yaml.NewDecoder(f).Decode(&fileProps)
	if err != nil {
		return nil, errors.WithMessage(err, "yaml decode")
	}

	flatten := bellows.Flatten(fileProps)
	config := map[string]string{}
	for key, value := range flatten {
		config[key] = fmt.Sprintf("%v", value)
	}
	return config, nil
}
