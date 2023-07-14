package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/integration-system/isp-kit/config"
	"github.com/integration-system/isp-kit/validator"
	"github.com/stretchr/testify/require"
)

type Subst struct {
	Value    string
	Anything string
}

type Example struct {
	Int  int `valid:"required"`
	Bool bool
	Dur  time.Duration
	Subst
	SetExplicit int `valid:"required"`
	SomeStruct  Subst
	EnvOverride string
	Slice       []Subst
}

func Test(t *testing.T) {
	require := require.New(t)

	err := os.Setenv("TEST_EnvOverride", "envoverride")
	require.NoError(err)
	err = os.Setenv("test_slice.[1].anything", "envoverride")
	require.NoError(err)

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
