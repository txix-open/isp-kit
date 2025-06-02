# Package `json`

Пакет `json` расширяет библиотеку [jsoniter](https://github.com/json-iterator/go) с дополнительными возможностями:

- Кастомное форматирование времени
- Автоматическое преобразование имен полей camelCase, если не указан json тэг у структуры
- Совместимость с API стандартной библиотеки `encoding/json`

## Functions

#### `Marshal(v any) ([]byte, error)`

Преобразовать переданный объект в массив байт json-формата.

#### `Unmarshal(data []byte, ptr any) error`

Преобразовать объект обратно из массива байт json-формата.

#### `NewEncoder(w io.Writer) *jsoniter.Encoder`

Создание потокового энкодера.

#### `NewDecoder(r io.Reader) *jsoniter.Decoder`

Создание потокового декодера.

#### `EncodeInto(w io.Writer, value any) error`

Оптимизированная запись в поток с автоматическим сбросом буфера

## Usage

### Default usage flow

```go
package main

import (
	"log"
	"time"

	"github.com/txix-open/isp-kit/json"
)

type Message struct {
	Text      string
	CreatedAt time.Time
}

func main() {
	msg := Message{
		Text:      "hi x.x",
		CreatedAt: time.Now().UTC(),
	}
	/* data: {"text":"hi x.x","createdAt":"2009-11-10T23:00:00Z"} */
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatal(err)
	}
}

```