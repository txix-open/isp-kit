package json_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/txix-open/isp-kit/json"
)

func TestFormat(t *testing.T) {
	t1 := time.Date(2020, 11, 3, 22, 2, 2, 893433646, time.UTC)
	fmt.Println(t1.Format(json.FullDateFormat))

	t2 := time.Date(2020, 11, 3, 22, 2, 2, 893433646, time.Local)
	fmt.Println(t2.Format(json.FullDateFormat))

	t3, err := time.Parse(json.FullDateFormat, "2020-11-03T22:02:02.8930984Z")
	if err != nil {
		t.Error(err)
	}
	t4 := time.Date(2020, 11, 3, 22, 2, 2, 893098400, time.UTC)
	if !t3.Equal(t4) {
		t.Errorf("%v != %v", t3, t4)
	}
}
