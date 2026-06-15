# Compogo Types

[![Go Reference](https://pkg.go.dev/badge/github.com/Compogo/types.svg)](https://pkg.go.dev/github.com/Compogo/types)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Коллекция типобезопасных и производительных структур данных на Go с использованием дженериков.

## Установка

```shell
go get github.com/Compogo/types
```

## Пакеты

### set — Множество

Типобезопасное множество на основе `map[T]struct{}`.

Пример:
```go
s := set.NewSet(1, 2, 3)
s.Add(4, 5)
intersection := s.Intersection(set.NewSet(2, 3, 6))
```

### ranger — Числовые диапазоны

Числовые диапазоны с включающими, исключающими и открытыми границами.

Пример:
```go
lower := ranger.NewInclusiveLowerBoundValue(0)
upper := ranger.NewExclusiveUpperBoundValue(100)
r, _ := ranger.NewRange(lower, upper)
r.Contains(50)
```

### mapper — Регистронезависимый словарь

Отображение для типов, реализующих `fmt.Stringer`.

Пример:
```go
type Status string
func (s Status) String() string { return string(s) }
m := mapper.NewMapper[Status](Status("active"), Status("inactive"))
status, _ := m.Get("ACTIVE")
```

### linker — Обобщённое отображение

Отображение с поддержкой нормализации ключей.

Пример:
```go
l := linker.NewLinker[string, int](
    linker.KeyStringNormalizer[int](),
    linker.Link("one", 1),
)
val, _ := l.Get("ONE")
```

### hash_slice — Слайс с поиском за O(1)

Сочетает свойства слайса и хэш-таблицы. Элементы уникальны.

Пример:
```go
hs, _ := hash_slice.NewHashSliceFrom("a", "b", "c")
idx := hs.IndexOf("b")
hs.Remove("b")
```

### emitter — Издатель-Подписчик

Асинхронная рассылка событий подписчикам.

Пример:
```go
e := emitter.NewEmitter[string]()
e.Subscribe(func(ctx context.Context, msg string) {
    fmt.Println(msg)
})
e.Emit(context.Background(), "hello")
e.Wait()
```

### errors — Общие ошибки

Централизованные ошибки проекта.

Пример:
```go
if errors.Is(err, errors.DoesNotExistError) {
// обработка
}
```

## Лицензия

MIT License

Copyright (c) 2026 Compogo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.