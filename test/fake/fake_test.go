package fake_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.txix.ru/isp/isp-kit/test/fake"
)

type SomeStruct struct {
	A string
	B bool
}

func Test(t *testing.T) {
	require := require.New(t)

	intValue := fake.It[int]()
	require.NotEmpty(intValue)

	stringSlice := fake.It[[]string]()
	require.NotEmpty(stringSlice)

	structSlice := fake.It[[]SomeStruct]()
	t.Log(structSlice)
	require.NotEmpty(structSlice)

	time := fake.It[time.Time]()
	require.False(time.IsZero())
}
