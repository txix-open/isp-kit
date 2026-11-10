# Package `file`

Пакет `file` предоставляет функциональность для настройки файлового вывода логов
с ротацией.

Ротация по мегабайтам выполняется через [Lumberjack](https://github.com/natefinch/lumberjack);
если задан `PartSizeKb`, используется собственный writer с байтово точным размером
части (для лимитов менее 1 МБ, например 218 КБ для VictoriaLogs).

Интегрируется с [zap](https://github.com/uber-go/zap) как кастомный Sink.

## Types

### Output

Структура представляет собой конфигурацию вывода, в которой можно указать путь к файлу логов,
размер части в килобайтах (`PartSizeKb`, приоритет) или максимальный размер файла в МБ
(`MaxSizeMb`) до ротации, максимальное количество старых лог-файлов и функцию
сжатия старых файлов.

## Functions

#### `ConfigToUrl(r Output) *url.URL`

Сконвертировать Output в URL для `zap`.

#### `NewFileWriter(r Output) io.WriteCloser`

Создать логгер `lumberjack` с ротацией.

## Usage

### Default usage flow

```go
package main

import (
	"log"

	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/file"
)

func main() {
	output := file.Output{
		File:       "app.log",
		MaxSizeMb:  100,  /* Ротация при 100MB */
		MaxBackups: 3,    /* Хранить 3 старых файла */
		Compress:   true, /* Сжимать .gz */
	}

	logger, err := log2.New(
		log2.WithLevel(log2.InfoLevel),
		log2.WithFileOutput(output),
	)
	if err != nil {
		log.Fatal(err)
	}
}

```

### Splitting into small parts

```go
output := file.Output{
	File:       "/var/log/app.log",
	PartSizeKb: 218, /* Приоритет над MaxSizeMb: часть не превышает 218 КБ */
	MaxBackups: 4,
	Compress:   false,
}

logger, err := log2.New(log2.WithFileOutput(output))
```