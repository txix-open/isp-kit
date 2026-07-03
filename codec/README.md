# Package `codec`

Пакет `codec` предоставляет реализации для кодирования и декодирования payload.

Поддерживаются:

* потоковое кодирование (`io.Writer` / `io.Reader`)
* кодирование/декодирование в память (`[]byte`)
* переиспользуемые реализации через pooling (для снижения аллокаций)

## Types

### Zstd

Реализация кодека на базе Zstandard.

#### `NewZstd(cfg ZstdConfig) *Zstd`

Создаёт новый экземпляр Zstd-кодека.

**ZstdConfig:**

* `Level` — уровень сжатия (`zstd.EncoderLevel`)

#### `(z *Zstd) EncodingType() string`

Возвращает тип кодирования:

```
"zstd"
```

## Encoding

### `(z *Zstd) Encode(w io.Writer) (io.WriteCloser, error)`

Создаёт потоковый энкодер.

Используется для записи данных в поток с последующим сжатием.

Пример:

```go
enc, err := codec.Encode(w)
if err != nil {
	return err
}
defer enc.Close()

_, _ = enc.Write(data)
```

### `(z *Zstd) EncodeBytes(data []byte) ([]byte, error)`

Кодирует байтовый срез в память.

Используется для случаев, когда не нужен streaming API.

## Decoding

### `(z *Zstd) Decode(body io.ReadCloser) (io.ReadCloser, error)`

Оборачивает входной поток декодером.

Возвращаемый `ReadCloser` автоматически декодирует данные при чтении.


### `(z *Zstd) DecodeBytes(data []byte) ([]byte, error)`

Декодирует данные из памяти.

## Internals

Пакет использует pooling энкодеров (`sync.Pool`) для снижения нагрузки на GC.

Основные оптимизации:

* повторное использование `zstd.Encoder`
* минимизация аллокаций при streaming encode
* разделение streaming и byte API

## Usage

### HTTP / middleware example

```go
codec := codec.Default

encoded, err := codec.EncodeBytes(data)
if err != nil {
	return err
}
```

### Decode example

```go
codec := codec.Default

decoded, err := codec.DecodeBytes(encoded)
if err != nil {
	return err
}
```