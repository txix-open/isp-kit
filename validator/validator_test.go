package validator_test

import (
	"maps"
	"strconv"
	"testing"

	"github.com/integration-system/isp-kit/validator"
	"github.com/stretchr/testify/require"
)

type Struct struct {
	Optional *SubStruct
}

type SubStruct struct {
	A string `validate:"required"`
}

func TestValidateNil(t *testing.T) {
	require := require.New(t)

	ok, details := validator.Default.Validate(Struct{})
	require.True(ok)
	require.Empty(details)

	ok, details = validator.Default.Validate(Struct{Optional: &SubStruct{}})
	expectedDetails := map[string]string{".Optional.A": "A is a required field"}
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
	expectedDetails := map[string]string{"[1].A": "A is a required field"}
	require.False(ok)
	require.EqualValues(expectedDetails, details)

	type s struct {
		Array []SubStruct
	}
	value := s{Array: arr}
	ok, details = validator.Default.Validate(value)
	expectedDetails = map[string]string{".Array[1].A": "A is a required field"}
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
	expectedDetails := map[string]string{"[key].A": "A is a required field"}
	require.False(ok)
	require.EqualValues(expectedDetails, details)
}

func TestCompositeDataType(t *testing.T) {
	require := require.New(t)
	type s struct {
		SomeStruct *Struct        `validate:"required"`
		Map        map[string]any `validate:"min=1"`
		Array      []map[int]any  `validate:"required,max=1"`
	}
	type a struct {
		Arr []*SubStruct `validate:"required,max=1"`
	}
	obj := &s{
		SomeStruct: &Struct{
			Optional: &SubStruct{},
		},
		Map: map[string]any{
			"key": &a{
				Arr: []*SubStruct{{A: "AAA"}, {}},
			},
			"sub": map[string]any{
				"map": &SubStruct{},
			},
		},
		Array: []map[int]any{
			{
				1: &a{
					Arr: []*SubStruct{{A: "AAA"}, {}},
				},
				2: map[string]any{
					"map": &SubStruct{},
				},
			},
			{
				5: "aa",
				6: "23213",
			},
		},
	}
	ok, details := validator.Default.Validate(obj)
	expectedDetails := map[string]string{
		".Array":                 "Array must contain at maximum 1 item",
		".Map[key].Arr":          "Arr must contain at maximum 1 item",
		".Map[sub][map].A":       "A is a required field",
		".SomeStruct.Optional.A": "A is a required field",
	}
	require.False(ok)
	require.EqualValues(expectedDetails, details)
}

func TestOneOf(t *testing.T) {
	require := require.New(t)
	type s struct {
		Int int    `validate:"oneof=5 22 913"`
		Str string `validate:"oneof='some string' 'str'"`
	}
	obj := &s{
		Int: 1,
		Str: "s",
	}
	ok, details := validator.Default.Validate(obj)
	expDetails := map[string]string{
		".Int": "Int must be one of [5 22 913]",
		".Str": "Str must be one of ['some string' 'str']",
	}
	require.False(ok)
	require.EqualValues(expDetails, details)
}

func genNestedStruct(depth int) *nested {
	if depth < 1 {
		return &nested{
			Map:   make(map[string]*nested),
			Array: make([]*nested, 0),
		}
	}
	obj := genNestedStruct(depth - 1)
	m := make(map[string]*nested)
	maps.Copy(m, obj.Map)
	m[strconv.Itoa(depth)] = obj
	return &nested{
		Struct: obj,
		Map:    m,
		Array:  append(obj.Array, obj),
	}
}

type nested struct {
	Struct *nested            `validate:"required" valid:"required"`
	Map    map[string]*nested `validate:"required" valid:"required"`
	Array  []*nested          `validate:"required" valid:"required"`
}

func BenchmarkValidatorV10(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		obj := genNestedStruct(5)
		for pb.Next() {
			validator.Default.Validate(obj)
		}
	})
}

func BenchmarkValidatorAsaskevich(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		obj := genNestedStruct(5)
		for pb.Next() {
			validator.Default.ValidateOld(obj)
		}
	})
}
