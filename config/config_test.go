package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/validator"
)

type Subst struct {
	Value    string
	Anything string
}

type Example struct {
	Subst

	Int         int `validate:"required"`
	Bool        bool
	Dur         time.Duration
	SetExplicit int `validate:"required"`
	SomeStruct  Subst
	EnvOverride string
	Slice       []Subst
}

func Test(t *testing.T) {
	t.Setenv("TEST_EnvOverride", "envoverride")
	t.Setenv("test_slice.[1].anything", "envoverride")
	require := require.New(t)

	cfg, err := config.New(
		config.WithExtraSource(config.NewYamlConfig("test_data/test.yml")),
		config.WithValidator(validator.Default),
		config.WithEnvPrefix("test_"),
	)
	require.NoError(err)

	actual := Example{}
	err = cfg.Read(&actual)
	require.Error(err)

	cfg.Set("SetExplicit", 1)

	actual = Example{}
	err = cfg.Read(&actual)
	require.NoError(err)

	expected := Example{
		Int:  7,
		Bool: true,
		Dur:  5 * time.Second,
		Subst: Subst{
			Value: "subst",
		},
		SetExplicit: 1,
		SomeStruct: Subst{
			Value: "someStruct",
		},
		EnvOverride: "envoverride",
		Slice: []Subst{{
			Value: "elem0",
		}, {
			Value:    "elem1",
			Anything: "envoverride",
		}, {
			Value: "elem2",
		}},
	}
	require.EqualValues(expected, actual)

	require.EqualValues(7, cfg.Optional().Int("INT", 5))

	_, err = cfg.Mandatory().Int("unknownKey")
	require.Error(err)

	_, err = cfg.Mandatory().Int("bool")
	require.Error(err)
}
