package throttling

import (
	"sync"
	"sync/atomic"
	"time"
)

type Throttling struct {
	beginDelay   time.Duration
	maxDelay     time.Duration
	incDelay     time.Duration
	multDelay    time.Duration
	maxCount     int64
	nextDelay    func(Throttling) time.Duration
	count        int64
	currentDelay time.Duration
	doneChan     chan struct{}
	doneChan2    <-chan struct{}

	sync.Mutex
	gotted bool
}

// NewThrottling - создание нового объекта
func NewThrottling(doneChan <-chan struct{}) *Throttling {
	return &Throttling{
		doneChan2: doneChan,
	}
}

// SetDelay - установить начальную задержку
func (o *Throttling) SetDelay(delay time.Duration) *Throttling {
	o.beginDelay = delay
	return o
}

// GetDelay - получить начальную задержку
func (o *Throttling) GetDelay() time.Duration {
	return o.beginDelay
}

// SetMaxDelay - установить максимальную задержку
func (o *Throttling) SetMaxDelay(delay time.Duration) *Throttling {
	o.maxDelay = delay
	return o
}

// GetMaxDelay - получить максимальную задержку
func (o *Throttling) GetMaxDelay() time.Duration {
	return o.maxDelay
}

// SetIncrement - установить коэффициент арифметической прогрессии
func (o *Throttling) SetIncrement(delay time.Duration) *Throttling {
	o.incDelay = delay
	return o
}

// GetIncrement - получить коэффициент арифметической прогрессии
func (o *Throttling) GetIncrement() time.Duration {
	return o.incDelay
}

// SetMultiplier - установить коэффициент геометрической прогрессии
func (o *Throttling) SetMultiplier(delay uint) *Throttling {
	o.multDelay = time.Duration(delay)
	return o
}

// GetMultiplier - получить коэффициент геометрической прогрессии
func (o *Throttling) GetMultiplier() time.Duration {
	return o.multDelay
}

// SetMaxCount - установить максимальное число итераций
func (o *Throttling) SetMaxCount(count int64) *Throttling {
	o.maxCount = count
	return o
}

// GetMaxCount - получить максимальное число итераций
func (o *Throttling) GetMaxCount() int64 {
	return o.maxCount
}

// SetDelayFunc - установить кастомную функцию расчета интервала задержки
func (o *Throttling) SetDelayFunc(f func(t Throttling) time.Duration) *Throttling {
	o.nextDelay = f
	return o
}

// GetDelayFunc - получить кастомную функцию расчета интервала задержки
func (o *Throttling) GetDelayFunc() func(t Throttling) time.Duration {
	return o.nextDelay
}

// Reset - возвращаем начальное состояние
func (o *Throttling) Reset() {
	o.Lock()
	defer o.Unlock()

	if o.gotted {
		o.currentDelay = o.beginDelay
	}

	o.gotted = false

}

// Throttling - функция возвращает канал, сообщение в котором появится после заданного через объект периода ожидания
//
//	если в канале будет true, то это означает, что работу можно заканчивать.
//	такое может произойти потому что превышено кол-во попыток (maxCount), либо пришла команда на завершение через doneChan
func (o *Throttling) Throttling() chan bool {
	ch := make(chan bool)
	atomic.AddInt64(&o.count, 1)
	o.Lock()
	o.gotted = true
	o.Unlock()

	go func(ch chan bool) {
		if o.maxCount != 0 && o.maxCount <= o.count {
			ch <- true
		}
		t := time.NewTicker(o.getCurrentDelay())

		defer func() {
			t.Stop()
		}()

		for {
			select {
			case <-o.doneChan:
				ch <- true
				return
			case <-o.doneChan2:
				ch <- true
				return
			case <-t.C:
				o.setNextCurrentDelay()
				close(ch)
				return
			}
		}
	}(ch)

	return ch
}

func (o *Throttling) getCurrentDelay() time.Duration {
	switch {
	case o.currentDelay == 0 && o.nextDelay != nil:
		o.currentDelay = o.nextDelay(*o)
	case o.currentDelay != 0:
		return o.currentDelay
	case o.beginDelay != 0:
		o.currentDelay = o.beginDelay
	default:
		o.currentDelay = time.Second
	}
	return o.currentDelay
}

func (o *Throttling) setNextCurrentDelay() {
	switch {
	case o.nextDelay != nil:
		o.currentDelay = o.nextDelay(*o)
	case o.incDelay > 0:
		o.currentDelay += o.incDelay
	case o.multDelay > 0:
		o.currentDelay *= o.multDelay
	}

	// проверяем, что не превысили максимальную задержку
	if o.maxDelay > 0 && o.currentDelay >= o.maxDelay {
		o.currentDelay = o.maxDelay
	}
}
