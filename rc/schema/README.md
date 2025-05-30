# Package `schema`

Пакет `schema` реализует генерацию и настройку JSON схем для структур Go с поддержкой кастомных генераторов и тегов валидации.

## Types

### CustomGenerator

Функция для кастомной генерации схемы по полю структуры.

### customSchema

Внутренняя структура для управления реестром кастомных генераторов.

**Methods:**

#### `Register(name string, f CustomGenerator)`

Регистрирует кастомный генератор под именем.

#### `Remove(name string)`

Удаляет кастомный генератор по имени.

### Schema

Псевдоним для `*github.com/txix-open/jsonschema.Schema`

### Generator

Генератор схемы.

**Methods:**

#### `NewGenerator()`

Создать и инициализировать генератор с настройками:

- `FieldNameReflector` — функция получения имени и признака обязательного поля,
- `FieldReflector` — функция установки свойств поля,
- `ExpandedStruct: true`,
- `DoNotReference: true`.

#### `Generate(obj any) Schema`

Сгенерировать JSON-схему для переданной структуры.

## Functions

### `func GetNameAndRequiredFlag(field reflect.StructField) (string, bool)`

- Возвращает имя поля (с учётом тега `json`) и флаг обязательности (наличие `validate:"required"`).
- Игнорирует неэкспортируемые поля.

### `func SetProperties(field reflect.StructField, s *jsonschema.Schema)`

- Извлекает из тегов поля `schema`, `default`, `validate`, `schemaGen`.
- Устанавливает свойства схемы: `Title`, `Description`, `Default`, валидационные ограничения.
- Если указан кастомный генератор (`schemaGen`), вызывает его.

### Валидация и ограничения

- Поддерживаются теги `validate` с опциями: `max`/`lte`, `min`/`gte`, `oneof`.
- Ограничения применяются к типам `string`, `integer`, `array`.
- Примеры:
  - `max=10` ограничение максимальной длины/значения,
  - `oneof='val1' 'val2'` — перечисление допустимых значений.

## Usage

### Default usage flow

```go
type MyConfig struct {
    Name string `json:"name" validate:"required,min=3" schema:"Имя,Описание поля"`
    Age  int    `json:"age" validate:"min=18,max=100" default:"18"`
}

gen := schema.NewGenerator()
jsonSchema := gen.Generate(MyConfig{})
// jsonSchema теперь содержит JSON-схему с ограничениями и описаниями
```
