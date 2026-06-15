// Package linker предоставляет обобщённое отображение (map) с поддержкой
// нормализации ключей и операций над множествами.
//
// Основные возможности:
//   - Нормализация ключей (регистронезависимость, приведение типов и т.д.)
//   - Функциональные опции для гибкой конфигурации
//   - Операции над множествами: Union, Intersection, Difference, SymmetricDifference
//   - Итерация с использованием iter.Seq2 (Go 1.23+)
//   - Клонирование и копирование
//
// Типичные сценарии использования:
//   - Регистронезависимые словари
//   - Кэши с предварительной обработкой ключей
//   - Маппинг с нормализацией входных данных
//
// Пример создания регистронезависимого линкера:
//
//	type User struct {
//	    ID   int
//	    Name string
//	}
//
//	linker := NewLinker[string, User](
//	    KeyStringNormalizer[User](), // регистронезависимость
//	    Link("john", User{ID: 1, Name: "John"}),
//	    Link("jane", User{ID: 2, Name: "Jane"}),
//	)
//
//	user, _ := linker.Get("JOHN") // возвращает пользователя John
//	linker.Has("jane")            // true
//
// Пример операций над множествами:
//
//	active := NewLinker[int, string](
//	    Link(1, "Alice"),
//	    Link(2, "Bob"),
//	)
//	admin := NewLinker[int, string](
//	    Link(2, "Bob"),
//	    Link(3, "Charlie"),
//	)
//
//	all := active.Union(admin)          // 1,2,3
//	both := active.Intersection(admin)  // 2
//	onlyActive := active.Difference(admin) // 1

package linker

import (
	"fmt"
	"iter"
	"maps"

	typeName "github.com/Compogo/tools/type_name"
	"github.com/Compogo/types/errors"
)

// Linker представляет собой обобщённое отображение (map) ключей типа T на значения типа I.
// Поддерживает:
//   - Нормализацию ключей (например, приведение к нижнему регистру)
//   - Функциональные опции для конфигурации
//   - Операции над множествами (Union, Intersection, Difference, SymmetricDifference)
//   - Итерацию через All()
//   - Клонирование и копирование
//
// Тип T должен быть сравниваемым (comparable), тип I может быть любым.
type Linker[T comparable, I any] struct {
	values    map[T]I
	typeName  string
	zeroValue I

	normalizer Normalizer[T]
}

// NewLinker создаёт новый Linker с опциональными настройками.
// Применяются следующие опции по умолчанию:
//   - Стандартный нормализатор (ключ без изменений)
//   - Опции переданные в вызове
//
// Пример:
//
//	type Key string
//	linker := NewLinker[Key, int](
//	    WithNormalizer[Key, int](func(k Key) Key { return Key(strings.ToLower(string(k))) }),
//	    Link(Key("one"), 1),
//	    Link(Key("two"), 2),
//	)
func NewLinker[T comparable, I any](options ...Option[T, I]) *Linker[T, I] {
	linker := &Linker[T, I]{
		values:    make(map[T]I),
		typeName:  fmt.Sprintf("Linker[%s, %s]", typeName.TypeName[T](), typeName.TypeName[I]()),
		zeroValue: *new(I),
	}

	options = append([]Option[T, I]{WithNormalizer[T, I](normalizerDefault)}, options...)
	for _, option := range options {
		option(linker)
	}

	return linker
}

// Add добавляет пару (ключ, значение) в Linker.
// Ключ перед добавлением нормализуется с помощью normalizer.
// Если ключ уже существует, значение будет перезаписано.
//
// Пример:
//
//	linker := NewLinker[string, int](KeyStringNormalizer[int]())
//	linker.Add("KEY", 42)
//	linker.Add("key", 100) // перезапишет значение для "key"
func (linker *Linker[T, I]) Add(key T, value I) {
	linker.values[linker.normalizer(key)] = value
}

// Get возвращает значение по ключу. Ключ нормализуется перед поиском.
//
// Возвращает:
//   - I — найденное значение
//   - error — ошибку DoesNotExistError, если ключ не найден
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1))
//	val, err := linker.Get("ONE")
//	if errors.Is(err, errors.DoesNotExistError) {
//	    // ключ не найден
//	}
func (linker *Linker[T, I]) Get(key T) (I, error) {
	if val, exists := linker.values[linker.normalizer(key)]; exists {
		return val, nil
	}

	return linker.zeroValue, fmt.Errorf("key %s %w for type %s", key, errors.DoesNotExistError, linker.typeName)
}

// GetOrDefault возвращает значение по ключу или значение по умолчанию d,
// если ключ не найден. Ключ нормализуется перед поиском.
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1))
//	val := linker.GetOrDefault("TWO", 0) // возвращает 0
func (linker *Linker[T, I]) GetOrDefault(key T, d I) I {
	if val, exists := linker.values[linker.normalizer(key)]; exists {
		return val
	}

	return d
}

// Has проверяет наличие ключа в Linker. Ключ нормализуется перед проверкой.
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1))
//	if linker.Has("ONE") {
//	    // ключ существует
//	}
func (linker *Linker[T, I]) Has(key T) bool {
	_, exists := linker.values[linker.normalizer(key)]
	return exists
}

// Remove удаляет ключ и связанное с ним значение из Linker.
// Ключ нормализуется перед удалением. Если ключ не найден, метод ничего не делает.
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1))
//	linker.Remove("ONE") // удаляет запись
func (linker *Linker[T, I]) Remove(key T) {
	delete(linker.values, linker.normalizer(key))
}

// Len возвращает количество элементов в Linker.
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1), Link("two", 2))
//	fmt.Println(linker.Len()) // 2
func (linker *Linker[T, I]) Len() int {
	return len(linker.values)
}

// Keys возвращает срез всех ключей, хранящихся в Linker.
// Порядок ключей не гарантируется.
//
// Пример:
//
//	for _, key := range linker.Keys() {
//	    fmt.Println(key)
//	}
func (linker *Linker[T, I]) Keys() []T {
	keys := make([]T, 0, len(linker.values))
	for k := range linker.values {
		keys = append(keys, k)
	}

	return keys
}

// Reset очищает Linker, удаляя все элементы.
//
// Пример:
//
//	linker.Reset()
//	fmt.Println(linker.Len()) // 0
func (linker *Linker[T, I]) Reset() {
	clear(linker.values)
}

// All возвращает итератор для обхода всех элементов Linker.
// Реализует интерфейс iter.Seq2, позволяя использовать range в цикле.
// Возвращает пары (ключ, значение) в нормализованном виде.
//
// Пример:
//
//	for key, value := range linker.All() {
//	    fmt.Printf("%v: %v\n", key, value)
//	}
//
// Итерацию можно прервать досрочно:
//
//	for key, value := range linker.All() {
//	    if someCondition(value) {
//	        break
//	    }
//	}
func (linker *Linker[T, I]) All() iter.Seq2[T, I] {
	return func(yield func(T, I) bool) {
		for key, value := range linker.values {
			if !yield(key, value) {
				return
			}
		}
	}
}

// Clone создаёт глубокую копию Linker.
// Нормализатор также копируется.
//
// Пример:
//
//	original := NewLinker[string, int](Link("one", 1))
//	copy := original.Clone()
//	copy.Add("two", 2) // original не изменится
func (linker *Linker[T, I]) Clone() *Linker[T, I] {
	return &Linker[T, I]{
		values:     maps.Clone(linker.values),
		typeName:   linker.typeName,
		zeroValue:  linker.zeroValue,
		normalizer: linker.normalizer,
	}
}

// HasAnd проверяет, содержатся ли ВСЕ переданные ключи в Linker.
// Возвращает true, только если каждый ключ присутствует.
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1), Link("two", 2))
//	linker.HasAnd("one", "two") // true
//	linker.HasAnd("one", "three") // false
func (linker *Linker[T, I]) HasAnd(items ...T) bool {
	for _, item := range items {
		if !linker.Has(item) {
			return false
		}
	}

	return true
}

// HasOr проверяет, содержится ли ХОТЯ БЫ ОДИН из переданных ключей в Linker.
// Возвращает true при первом же найденном ключе.
//
// Пример:
//
//	linker := NewLinker[string, int](Link("one", 1), Link("two", 2))
//	linker.HasOr("three", "one") // true
//	linker.HasOr("three", "four") // false
func (linker *Linker[T, I]) HasOr(items ...T) bool {
	for _, item := range items {
		if linker.Has(item) {
			return true
		}
	}

	return false
}

// Intersection возвращает новый Linker, содержащий только те пары (ключ, значение),
// ключи которых присутствуют в обоих линкерах. Значения берутся из линкера,
// который используется для итерации (того, который меньше по размеру).
//
// Пример:
//
//	a := NewLinker[string, int](Link("one", 1), Link("two", 2))
//	b := NewLinker[string, int](Link("two", 20), Link("three", 3))
//	c := a.Intersection(b) // содержит {"two": 2} или {"two": 20} (зависит от размера)
func (linker *Linker[T, I]) Intersection(l *Linker[T, I]) *Linker[T, I] {
	resultLinker := NewLinker[T, I]()

	var linkerForRange, linkerForCondition *Linker[T, I]
	if linker.Len() > l.Len() {
		linkerForRange = l
		linkerForCondition = linker
	} else {
		linkerForRange = linker
		linkerForCondition = l
	}

	for key, val := range linkerForRange.All() {
		if linkerForCondition.Has(key) {
			resultLinker.Add(key, val)
		}
	}

	return resultLinker
}

// SymmetricDifference возвращает новый Linker, содержащий элементы,
// которые присутствуют либо в текущем линкере, либо в переданном,
// но не в обоих одновременно.
//
// Пример:
//
//	a := NewLinker[string, int](Link("one", 1), Link("two", 2))
//	b := NewLinker[string, int](Link("two", 20), Link("three", 3))
//	c := a.SymmetricDifference(b) // содержит {"one": 1, "three": 3}
func (linker *Linker[T, I]) SymmetricDifference(l *Linker[T, I]) *Linker[T, I] {
	var resultLinker, linkerForRange *Linker[T, I]
	if linker.Len() > l.Len() {
		resultLinker = linker.Clone()
		linkerForRange = l
	} else {
		resultLinker = l.Clone()
		linkerForRange = linker
	}

	for key, val := range linkerForRange.All() {
		if !resultLinker.Has(key) {
			resultLinker.Add(key, val)
		} else {
			resultLinker.Remove(key)
		}
	}

	return resultLinker
}

// Difference возвращает новый Linker, содержащий элементы текущего линкера,
// ключи которых отсутствуют в переданном линкере.
//
// Пример:
//
//	a := NewLinker[string, int](Link("one", 1), Link("two", 2))
//	b := NewLinker[string, int](Link("two", 20), Link("three", 3))
//	c := a.Difference(b) // содержит {"one": 1}
func (linker *Linker[T, I]) Difference(l *Linker[T, I]) *Linker[T, I] {
	if linker.Len() == 0 {
		return NewLinker[T, I]()
	}

	newLinker := linker.Clone()
	for key := range l.All() {
		newLinker.Remove(key)
	}
	return newLinker
}

// Union возвращает объединение текущего линкера и переданного.
// Использует стратегию Optimization (см. UnionByStrategy).
//
// Пример:
//
//	a := NewLinker[string, int](Link("one", 1))
//	b := NewLinker[string, int](Link("two", 2))
//	c := a.Union(b) // содержит {"one": 1, "two": 2}
//
// При конфликте ключей значение берётся из линкера, который был клонирован
// (зависит от стратегии).
func (linker *Linker[T, I]) Union(l *Linker[T, I]) *Linker[T, I] {
	return linker.UnionByStrategy(l, Optimization)
}

// UnionByStrategy возвращает объединение линкеров с использованием указанной стратегии.
// Доступные стратегии:
//   - Optimization (умный выбор — клонируется больший линкер)
//   - CurrentLinker (клонируется текущий линкер, добавляются элементы из l)
//   - IncomingLinker (клонируется линкер l, добавляются элементы из текущего)
//
// При конфликте ключей значение берётся из клонируемого линкера.
//
// Пример:
//
//	a := NewLinker[string, int](Link("one", 1))
//	b := NewLinker[string, int](Link("one", 100))
//	c := a.UnionByStrategy(b, CurrentLinker) // c["one"] == 1
//	d := a.UnionByStrategy(b, IncomingLinker) // d["one"] == 100
func (linker *Linker[T, I]) UnionByStrategy(l *Linker[T, I], strategy Strategy) *Linker[T, I] {
	switch strategy {
	case CurrentLinker:
		resultSet := linker.Clone()
		Copy(resultSet, l)
		return resultSet
	case IncomingLinker:
		resultSet := l.Clone()
		Copy(resultSet, linker)
		return resultSet
	default:
		return linker.unionOptimization(l)
	}
}

// unionOptimization — внутренняя реализация объединения с оптимизацией по размеру.
func (linker *Linker[T, I]) unionOptimization(l *Linker[T, I]) *Linker[T, I] {
	var resultLinker, linkerForRange *Linker[T, I]

	if linker.Len() > l.Len() {
		resultLinker = linker.Clone()
		linkerForRange = l
	} else {
		resultLinker = l.Clone()
		linkerForRange = linker
	}

	for key, val := range linkerForRange.All() {
		resultLinker.Add(key, val)
	}

	return resultLinker
}

// Replace заменяет содержимое текущего линкера на содержимое переданного.
// Текущий линкер очищается, после чего в него копируются все элементы из l.
// Нормализатор текущего линкера не изменяется.
//
// Пример:
//
//	a := NewLinker[string, int](Link("one", 1))
//	b := NewLinker[string, int](Link("two", 2))
//	a.Replace(b) // a теперь содержит {"two": 2}
func (linker *Linker[T, I]) Replace(l *Linker[T, I]) {
	linker.Reset()
	maps.Copy(linker.values, l.values)
}

// Copy копирует все элементы из src в dst.
// Если ключ уже существует в dst, он будет перезаписан.
//
// Пример:
//
//	dst := NewLinker[string, int]()
//	src := NewLinker[string, int](Link("one", 1))
//	Copy(dst, src) // dst содержит {"one": 1}
func Copy[T comparable, I any](dst *Linker[T, I], src *Linker[T, I]) {
	for key, val := range src.values {
		dst.Add(key, val)
	}
}
