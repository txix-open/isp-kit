package validator_test

import (
	"testing"

	"github.com/integration-system/isp-kit/validator"
	"github.com/stretchr/testify/require"
)

type Struct struct {
	Optional *SubStruct
}

type SubStruct struct {
	A string `valid:"required"`
}

func TestValidateNil(t *testing.T) {
	require := require.New(t)

	ok, details := validator.Default.Validate(Struct{})
	require.True(ok)
	require.Empty(details)

	ok, details = validator.Default.Validate(Struct{Optional: &SubStruct{}})
	expectedDetails := map[string]string{"Optional.A": "non zero value required"}
	require.False(ok)
	require.EqualValues(expectedDetails, details)
}

func TestValidateArray(t *testing.T) {
	require := require.New(t)

	arr := []SubStruct{{
		A: "1",
	}, {
		A: "",
	}}
	ok, details := validator.Default.Validate(arr)
	expectedDetails := map[string]string{"1.A": "non zero value required"}
	require.False(ok)
	require.EqualValues(expectedDetails, details)

	type s struct {
		Array []SubStruct
	}
	value := s{Array: arr}
	ok, details = validator.Default.Validate(value)
	expectedDetails = map[string]string{"Array.1.A": "non zero value required"}
	require.False(ok)
	require.EqualValues(expectedDetails, details)
}

func TestMap(t *testing.T) {
	require := require.New(t)

	m := map[string]SubStruct{
		"key":  {A: ""},
		"key2": {A: "2"},
	}
	ok, details := validator.Default.Validate(m)
	expectedDetails := map[string]string{"key.A": "non zero value required"}
	require.False(ok)
	require.EqualValues(expectedDetails, details)
}
