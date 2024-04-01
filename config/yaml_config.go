package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"gopkg.in/yaml.v3"
)

type Source interface {
	Config() (map[string]string, error)
}

type YamlFileSource struct {
	file string
}

func NewYamlConfig(file string) YamlFileSource {
	return YamlFileSource{file: file}
}

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
