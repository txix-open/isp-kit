# throttling

Библиотека реализует неблокирующую задержку. Т.е. time.Sleep, который можно прервать. 

Основная функция возвращает канал, сообщение в котором появится после заданного через объект периода ожидания. Если в канале будет true, то это означает, что работу можно заканчивать. Такое может произойти потому что превышено кол-во попыток (maxCount), либо пришла команда на завершение через doneChan


- doneChan     <-chan struct{} - ОБЯЗАТЕЛЬНЫЙ. Канал, который может прекратить выполнение задержки. Если его закрыть, то в канал, который возвращает основная функция, вернется true 
- beginDelay   time.Duration - Начальная задержка
- maxDelay     time.Duration - Максимальная задержка
- incDelay     time.Duration - Сколько добавить (сложение) после каждой неудачной попытки 
- multDelay    time.Duration - На сколько увеличить (умножение) после каждой неудачной попытки
- maxCount     int64 - Максимальное кол-во потыток. Когда оно будет превышено, то в канал, который возвращает основная функция, верется true
- nextDelay    func(Throttling) time.Duration - Функция для расчета следующей задержки 


## Примеры

Создаем правила для работы библиотеки. Начальная задержка будет 2 секунды, она не будет меняться. Функция doSomething будет выполнена не более 2 раз. Если она будет возвращаться ошибку.

```go
	tr := NewThrottling(doneChan).SetDelay(2 * time.Second).SetMaxCount(2)
	for {
		if err := doSomething(last, 2); err == nil {
			break
		}
		if <-tr.Throttling() {
			return
		}
	}
```
## Функции и методы

### NewThrottling - создание нового объекта
```go
func NewThrottling(doneChan chan struct{}) *Throttling
```

### SetDelay - установить начальную задержку
```go
func (o *Throttling) SetDelay(delay time.Duration) *Throttling
```

### GetDelay - получить начальную задержку
```go
func (o *Throttling) GetDelay(delay time.Duration) time.Duration
```

### SetMaxDelay - установить максимальную задержку
```go
func (o *Throttling) SetMaxDelay(delay time.Duration) *Throttling
```

### GetMaxDelay - получить максимальную задержку
```go
func (o *Throttling) GetMaxDelay(delay time.Duration) time.Duration
```

### SetIncrement - установить коэффициент арифметической прогрессии
```go
func (o *Throttling) SetIncrement(delay time.Duration) *Throttling
```

### GetIncrement - получить коэффициент арифметической прогрессии
```go
func (o *Throttling) GetIncrement(delay time.Duration) time.Duration
```

### SetMultiplier - установить коэффициент геометрической прогрессии
```go
func (o *Throttling) SetMultiplier(delay time.Duration) *Throttling
```

### GetMultiplier - получить коэффициент геометрической прогрессии
```go
func (o *Throttling) GetMultiplier(delay time.Duration) time.Duration
```

### SetMaxCount - установить максимальное число итераций
```go
func (o *Throttling) SetMaxCount(count int64) *Throttling
```

### GetMaxCount - получить максимальное число итераций
```go
func (o *Throttling) GetMaxCount(count int64) int64
```

### SetDelayFunc - установить кастомную функцию расчета интервала задержки
```go
func (o *Throttling) SetDelayFunc(f func(t Throttling) time.Duration) *Throttling
```

### GetDelayFunc - получить кастомную функцию расчета интервала задержки
```go
func (o *Throttling) GetDelayFunc(f func(t Throttling) time.Duration) func(t Throttling) time.Duration
```

### Throttling - функция возвращает канал, сообщение в котором появится после заданного через объект периода ожидания если в канале будет true, то это означает, что работу можно заканчивать. Такое может произойти потому что превышено кол-во попыток (maxCount), либо пришла команда на завершение через doneChan
```go
func (o *Throttling) Throttling() chan bool
```
