package throttling

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func doSomething(last, diff int64) (bool, string) {
	n := time.Now().Unix()
	fmt.Println("\tDo something at " + time.Now().Format(time.RFC3339Nano))
	if last != 0 {
		if n-last != diff {
			return false, strconv.FormatInt(n, 10) + "-" + strconv.FormatInt(last, 10) + "!=" + strconv.FormatInt(diff, 10)
		}
	}
	return true, ""

}

func Test_One(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetMaxCount(2).SetDelay(2 * time.Second)
	// tr := Throttling{MaxCount: 2, DoneChan: d, BeginDelay: 2 * time.Second}
	var last int64
	for {
		if ok, errStr := doSomething(last, 2); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		last = time.Now().Unix()
		ok := <-tr.Throttling()
		if ok {
			return
		}
	}
}

func Test_Two(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d)
	c := 0
	var last int64
	for {
		if ok, errStr := doSomething(last, 1); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		last = time.Now().Unix()
		ok := <-tr.Throttling()
		if ok {
			return
		}
		if c == 2 {
			close(d)
		}
		c++
	}
}

func Test_Three(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetDelay(time.Second).SetIncrement(time.Second).SetMaxCount(4)
	// tr := Throttling{MaxCount: 4, DoneChan: d, BeginDelay: time.Second, IncDelay: time.Second}
	var (
		last int64
		diff int64
	)
	for {
		if ok, errStr := doSomething(last, diff); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		diff += 1
		last = time.Now().Unix()
		ok := <-tr.Throttling()
		if ok {
			return
		}
	}
}

func Test_Four(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetDelay(time.Second).SetMultiplier(2).SetMaxCount(5)
	// tr := Throttling{MaxCount: 5, DoneChan: d, BeginDelay: time.Second, MultDelay: 2}
	var (
		last int64
	)
	diffs := []int64{1, 1, 2, 4, 8}
	count := 0
	for {
		if ok, errStr := doSomething(last, diffs[count]); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		count++
		last = time.Now().Unix()
		ok := <-tr.Throttling()
		if ok {
			return
		}
	}
}

func Test_Five(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetDelay(time.Second).SetMultiplier(2).SetMaxDelay(7 * time.Second).SetMaxCount(6)
	// tr := Throttling{MaxCount: 6, DoneChan: d, BeginDelay: time.Second, MultDelay: 2, MaxDelay: 7 * time.Second}

	var (
		last int64
	)
	diffs := []int64{1, 1, 2, 4, 7, 7}
	count := 0
	for {
		if ok, errStr := doSomething(last, diffs[count]); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		count++
		last = time.Now().Unix()
		if <-tr.Throttling() {
			return
		}
	}
}

func nextDelay(o Throttling) time.Duration {
	if o.count%2 == 0 {
		return time.Second
	}
	return 2 * time.Second
}

func Test_Six(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetDelay(time.Second).SetMaxCount(8).SetMaxDelay(7 * time.Second).SetDelayFunc(nextDelay)
	// tr := Throttling{MaxCount: 8, DoneChan: d, BeginDelay: time.Second, MaxDelay: 7 * time.Second}
	// tr.NextDelay = nextDelay

	var (
		last int64
	)
	diffs := []int64{2, 2, 2, 1, 2, 1, 2, 1}
	count := 0
	for {
		if ok, errStr := doSomething(last, diffs[count]); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		count++
		last = time.Now().Unix()
		if <-tr.Throttling() {
			return
		}
	}

}

func Test_Seven(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetDelay(2 * time.Second).SetMaxCount(6)
	// tr := Throttling{MaxCount: 6, DoneChan: d, BeginDelay: 2 * time.Second}

	var last int64
	for i := 0; i < 5; i++ {
		if ok, errStr := doSomething(last, 2); !ok {
			t.Error("Ошибка:" + errStr)
			return
		}
		last = time.Now().Unix()
		<-tr.Throttling()
	}
}

func TestReset(t *testing.T) {
	t.Parallel()
	d := make(chan struct{})
	tr := NewThrottling(d).SetDelay(time.Second).SetMaxDelay(5 * time.Second).SetMultiplier(2)

	for i := 1; i <= 20; i++ {
		switch i {
		case 3, 13:
			if tr.getCurrentDelay().Seconds() != 1.0 {
				t.Error("Ошибка 3")
			}
			<-tr.Throttling()
		case 4, 14:
			if tr.getCurrentDelay().Seconds() != 2.0 {
				t.Error("Ошибка 4")
			}
			<-tr.Throttling()
		case 5, 15:
			if tr.getCurrentDelay().Seconds() != 4.0 {
				t.Error("Ошибка 5")
			}
			<-tr.Throttling()
		case 9, 19:
			if tr.getCurrentDelay().Seconds() != 1.0 {
				t.Error("Ошибка 9")
			}
			<-tr.Throttling()
		default:
			tr.Reset()
		}
	}

}

func sec() int64 {
	return time.Now().Unix()
}
