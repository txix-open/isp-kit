package rc_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/rc"
)

type noneValidation struct {
}

func (n noneValidation) ValidateToError(value any) error {
	return nil
}

func TestConfig_Upgrade(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	override := `{"a": {"a": 1}}`
	config := rc.New(noneValidation{}, []byte(override))

	type cfgType struct {
		A struct {
			A int
		}
		B int
	}

	cfg1 := `{"a": {"a": 2}, "b": 2}`
	expectedNewCfg := cfgType{
		A: struct {
			A int
		}{1},
		B: 2,
	}
	expectedPrevCfg := cfgType{}
	newCfg := cfgType{}
	prevCfg := cfgType{}
	err := config.Upgrade([]byte(cfg1), &newCfg, &prevCfg)
	require.NoError(err)
	require.EqualValues(expectedNewCfg, newCfg)
	require.EqualValues(expectedPrevCfg, prevCfg)

	cfg2 := `{"a": {"a": 4}, "b": 3}`
	expectedNewCfg = cfgType{
		A: struct {
			A int
		}{1},
		B: 3,
	}
	expectedPrevCfg = cfgType{
		A: struct {
			A int
		}{1},
		B: 2,
	}
	newCfg = cfgType{}
	prevCfg = cfgType{}
	err = config.Upgrade([]byte(cfg2), &newCfg, &prevCfg)
	require.NoError(err)
	require.EqualValues(expectedNewCfg, newCfg)
	require.EqualValues(expectedPrevCfg, prevCfg)
}

func TestConfig_Delim(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	override := []byte(`{"map.key.1": "overridden.value"}`)
	config := rc.New(noneValidation{}, override)
	cfg := []byte(`{"map.key.1":"map.value","map.key.2":"map.value"}`)
	expectedNewCfg := map[string]string{
		"map.key.1": "overridden.value",
		"map.key.2": "map.value",
	}
	expectedPrevCfg := make(map[string]string)
	newCfg := make(map[string]string)
	prevCfg := make(map[string]string)
	err := config.Upgrade(cfg, &newCfg, &prevCfg)
	require.NoError(err)
	require.EqualValues(expectedNewCfg, newCfg)
	require.EqualValues(expectedPrevCfg, prevCfg)
}
