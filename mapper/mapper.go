// Package mapper предоставляет регистронезависимое отображение для типов,
// реализующих fmt.Stringer.
//
// Основные возможности:
//   - Автоматическая нормализация ключей (to lower)
//   - Безопасный доступ с возвратом ошибки
//   - Итерация с использованием iter.Seq2 (Go 1.23+)
//   - Поддержка пользовательских типов через дженерики
//
// Типичные сценарии использования:
//   - Создание enum-подобных структур
//   - Регистронезависимые словари (например, HTTP-заголовки, названия цветов)
//   - Кэширование строковых представлений объектов
//
// Пример определения типа, реализующего fmt.Stringer:
//
//	type Status string
//	func (s Status) String() string { return string(s) }
//
// Пример создания и использования маппера:
//
//	mapper := NewMapper[Status](Status("active"), Status("inactive"))
//
//	status, err := mapper.Get("ACTIVE")   // возвращает Status("active")
//	exists := mapper.HasByKey("Inactive") // true
//	mapper.RemoveByValue(Status("PENDING"))
//
// Пример итерации:
//
//	for key, value := range mapper.All() {
//	    fmt.Printf("Ключ: %s, Значение: %v\n", key, value)
//	}

package mapper

import (
	"fmt"
	"iter"
	"strings"

	typeName "github.com/Compogo/tools/type_name"
	"github.com/Compogo/types/errors"
)

// Mapper представляет собой регистронезависимое отображение строковых ключей
// на значения типа T, который должен реализовывать интерфейс fmt.Stringer.
//
// Mapper обеспечивает:
//   - Автоматическое приведение ключей к нижнему регистру
//   - Безопасное получение значений с возвратом ошибки
//   - Итерацию по всем элементам через All()
//   - Типобезопасность через дженерики
//
// Поля структуры не экспортируются, используйте методы для работы с Mapper.
type Mapper[T fmt.Stringer] struct {
	items     map[string]T
	typeName  string
	zeroValue T
}

// NewMapper создаёт новый Mapper с опциональным начальным набором значений.
// Все переданные значения добавляются в mapper с ключами, полученными через
// их метод String() (приведёнными к нижнему регистру).
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	colors := NewMapper[Color](Color("red"), Color("green"), Color("blue"))
func NewMapper[T fmt.Stringer](values ...T) *Mapper[T] {
	mapper := &Mapper[T]{
		items:     make(map[string]T),
		typeName:  fmt.Sprintf("Mapper[%s]", typeName.TypeName[T]()),
		zeroValue: *new(T),
	}

	mapper.Add(values...)

	return mapper
}

// Add добавляет одно или несколько значений в Mapper.
// Для каждого значения ключом становится результат value.String(),
// приведённый к нижнему регистру. Если значение с таким ключом уже существует,
// оно будет перезаписано.
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color]()
//	mapper.Add(Color("red"), Color("blue"))
//	mapper.Add(Color("RED")) // перезапишет Color("red")
func (mapper *Mapper[T]) Add(values ...T) {
	for _, value := range values {
		mapper.items[strings.ToLower(value.String())] = value
	}
}

// Get возвращает значение по строковому ключу. Поиск выполняется регистронезависимо:
// ключи "Key", "KEY", "key" считаются эквивалентными.
//
// Возвращает:
//   - T — найденное значение
//   - error — ошибку DoesNotExistError, если ключ не найден
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color](Color("red"), Color("green"))
//	value, err := mapper.Get("RED")
//	if errors.Is(err, errors.DoesNotExistError) {
//	    // ключ не найден
//	}
func (mapper *Mapper[T]) Get(key string) (T, error) {
	if val, exists := mapper.items[strings.ToLower(key)]; exists {
		return val, nil
	}

	return mapper.zeroValue, fmt.Errorf("key %s %w for type %s", key, errors.DoesNotExistError, mapper.typeName)
}

// Keys возвращает срез всех ключей, хранящихся в Mapper.
// Ключи возвращаются в нижнем регистре (нормализованная форма).
// Порядок ключей не гарантируется (зависит от итерации по map).
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color](Color("red"), Color("green"), Color("blue"))
//	for _, key := range mapper.Keys() {
//	    fmt.Println(key) // "red", "green", "blue" в произвольном порядке
//	}
func (mapper *Mapper[T]) Keys() []string {
	keys := make([]string, 0, len(mapper.items))
	for k := range mapper.items {
		keys = append(keys, k)
	}

	return keys
}

// HasByValue проверяет наличие значения в Mapper по его строковому представлению.
// Метод регистронезависим — вызывает value.String() и приводит результат к нижнему регистру.
// Эквивалентен вызову HasByKey(value.String()).
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color](Color("red"), Color("green"))
//	if mapper.HasByValue(Color("RED")) {
//	    // значение присутствует
//	}
func (mapper *Mapper[T]) HasByValue(value T) bool {
	return mapper.HasByKey(value.String())
}

// HasByKey проверяет наличие ключа в Mapper. Поиск выполняется регистронезависимо.
//
// Пример:
//
//	type Status string
//	func (s Status) String() string { return string(s) }
//
//	mapper := NewMapper[Status](Status("active"), Status("inactive"))
//	if mapper.HasByKey("ACTIVE") {
//	    // ключ существует
//	}
func (mapper *Mapper[T]) HasByKey(key string) bool {
	_, exists := mapper.items[strings.ToLower(key)]
	return exists
}

// RemoveByValue удаляет значение из Mapper по его строковому представлению.
// Вызывает value.String() для получения ключа. Если значение не найдено,
// метод ничего не делает (не вызывает панику).
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color](Color("red"), Color("green"), Color("blue"))
//	mapper.RemoveByValue(Color("RED")) // удалит Color("red")
func (mapper *Mapper[T]) RemoveByValue(value T) {
	mapper.RemoveByKey(value.String())
}

// RemoveByKey удаляет значение из Mapper по строковому ключу.
// Поиск выполняется регистронезависимо. Если ключ не найден,
// метод ничего не делает.
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color](Color("red"), Color("green"), Color("blue"))
//	mapper.RemoveByKey("RED")   // удалит Color("red")
//	mapper.RemoveByKey("green") // удалит Color("green")
func (mapper *Mapper[T]) RemoveByKey(key string) {
	delete(mapper.items, strings.ToLower(key))
}

// All возвращает итератор для обхода всех элементов Mapper.
// Реализует интерфейс iter.Seq2, позволяя использовать range в цикле.
// Возвращает пары (ключ, значение), где ключ приведён к нижнему регистру.
//
// Пример:
//
//	type Color string
//	func (c Color) String() string { return string(c) }
//
//	mapper := NewMapper[Color](Color("red"), Color("green"), Color("blue"))
//	for key, value := range mapper.All() {
//	    fmt.Printf("%s: %s\n", key, value)
//	}
//	// Вывод (порядок может быть разным):
//	// red: red
//	// green: green
//	// blue: blue
//
// Итерацию можно прервать досрочно:
//
//	for key, value := range mapper.All() {
//	    if value == Color("green") {
//	        break
//	    }
//	}
func (mapper *Mapper[T]) All() iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for key, value := range mapper.items {
			if !yield(key, value) {
				return
			}
		}
	}
}
