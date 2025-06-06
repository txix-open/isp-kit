# Package `file`

Пакет `file` предоставляет функциональность для настройки файлового вывода логов с ротацией
через [Lumberjack](https://github.com/natefinch/lumberjack).  
Интегрируется с [zap](https://github.com/uber-go/zap) как кастомный Sink.

## Types

### Output

Структура представляет собой конфигурацию вывода, в которой можно указать путь к файлу логов, максимальный размер
файла (MB) до ротации, максимальный срок хранения файлов в днях, максимальное количество старых лог-файлов и функцию
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