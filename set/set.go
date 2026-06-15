package set

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

// numberOfElementsToRecreate — пороговое значение, при превышении которого
// операции добавления и слияния создают новое выделение памяти вместо
// итеративного расширения. Оптимизировано для уменьшения количества реаллокаций.
const numberOfElementsToRecreate = 5

// Set представляет собой типобезопасное множество, реализованное на основе
// встроенной мапы (map[T]struct{}). Поддерживает все стандартные операции
// над множествами: добавление, удаление, проверку вхождения, объединение,
// пересечение, разность и симметрическую разность.
//
// Нулевое значение Set является nil-множеством (пустым, но не выделенным).
// Большинство методов корректно работают с nil-получателем, возвращая пустые
// значения или выполняя операции без паники.
type Set[T comparable] map[T]struct{}

// exists — пустая структура, используемая для экономии памяти.
// В Go map[T]struct{} не занимает дополнительной памяти под значение,
// что делает эту реализацию Set одной из самых эффективных.
var exists = struct{}{}

// NewSet создает и возвращает новое множество, содержащее переданные элементы items.
// Если элементы не переданы, возвращается пустое выделенное множество.
//
// Пример:
//
//	s := NewSet(1, 2, 3)
//	empty := NewSet[string]()
func NewSet[T comparable](items ...T) Set[T] {
	s := Set[T]{}

	s.Add(items...)

	return s
}

// Add добавляет один или несколько элементов в множество.
// Если items равен nil или длина слайса элементов равна 0, метод ничего не делает.
// При добавлении большого количества элементов (>= numberOfElementsToRecreate)
// автоматически выполняет реаллокацию для оптимизации производительности.
//
// Пример:
//
//	s := NewSet[int]()
//	s.Add(1, 2, 3) // в s находятся {1, 2, 3}
func (s *Set[T]) Add(items ...T) {
	if s == nil || len(items) == 0 {
		return
	}

	if len(items) >= numberOfElementsToRecreate {
		newSet := make(map[T]struct{}, s.Len()+len(items))

		if *s != nil {
			for item := range *s {
				newSet[item] = exists
			}
		}

		for _, item := range items {
			newSet[item] = exists
		}

		*s = newSet
		return
	}

	if *s == nil {
		*s = NewSet[T]()
	}

	for _, item := range items {
		(*s)[item] = exists
	}
}

// Merge добавляет все элементы из переданного множества set в текущее.
// Если текущее множество nil, оно заменяется на копию set.
// При слиянии больших множеств выполняется оптимизированное копирование.
//
// Пример:
//
//	a := NewSet(1, 2)
//	b := NewSet(2, 3)
//	a.Merge(b) // a содержит {1, 2, 3}
func (s *Set[T]) Merge(set Set[T]) {
	if s == nil || set.Len() == 0 {
		return
	}

	if !s.isAllocated() {
		*s = set.Clone()
		return
	}

	if set.Len() > numberOfElementsToRecreate {
		newSet := make(map[T]struct{}, s.Len()+set.Len())
		maps.Copy(newSet, *s)
		*s = newSet
	}

	maps.Copy(*s, set)
}

// Contains проверяет, содержится ли элемент item в множестве.
// Возвращает true, если элемент присутствует, и false в противном случае.
// Для nil-множества всегда возвращает false.
//
// Пример:
//
//	s := NewSet(1, 2)
//	s.Contains(1) // true
//	s.Contains(3) // false
func (s *Set[T]) Contains(item T) bool {
	if !s.isAllocated() {
		return false
	}

	_, isExists := (*s)[item]
	return isExists
}

// ContainsAnd проверяет, содержатся ли ВСЕ элементы items в множестве.
// Возвращает true, только если каждый переданный элемент присутствует.
//
// Пример:
//
//	s := NewSet(1, 2, 3)
//	s.ContainsAnd(1, 2) // true
//	s.ContainsAnd(1, 4) // false
func (s *Set[T]) ContainsAnd(items ...T) bool {
	for _, item := range items {
		if !s.Contains(item) {
			return false
		}
	}

	return true
}

// ContainsOr проверяет, содержится ли ХОТЯ БЫ ОДИН из элементов items в множестве.
// Возвращает true при первом же найденном элементе.
//
// Пример:
//
//	s := NewSet(1, 2, 3)
//	s.ContainsOr(4, 5, 1) // true
//	s.ContainsOr(4, 5, 6) // false
func (s *Set[T]) ContainsOr(items ...T) bool {
	for _, item := range items {
		if s.Contains(item) {
			return true
		}
	}

	return false
}

// Intersection возвращает новое множество, содержащее элементы,
// присутствующие одновременно в текущем множестве и в множестве set.
// Операция выполняется итерацией по меньшему множеству для оптимальной производительности.
//
// Пример:
//
//	a := NewSet(1, 2, 3)
//	b := NewSet(2, 3, 4)
//	c := a.Intersection(b) // c содержит {2, 3
func (s *Set[T]) Intersection(set Set[T]) Set[T] {
	resultSet := NewSet[T]()

	var setForRange, setForCondition Set[T]
	//range for largest
	if s.Len() > set.Len() {
		setForRange = set
		setForCondition = *s
	} else {
		setForRange = *s
		setForCondition = set
	}

	for item := range setForRange {
		if setForCondition.Contains(item) {
			resultSet.Add(item)
		}
	}

	return resultSet
}

// SymmetricDifference возвращает новое множество, содержащее элементы,
// которые присутствуют либо в текущем множестве, либо в множестве set,
// но не в обоих одновременно (симметрическая разность).
//
// Пример:
//
//	a := NewSet(1, 2, 3)
//	b := NewSet(3, 4, 5)
//	c := a.SymmetricDifference(b) // c содержит {1, 2, 4, 5}
func (s *Set[T]) SymmetricDifference(set Set[T]) Set[T] {
	if !s.isAllocated() && !set.isAllocated() {
		return NewSet[T]()
	}

	if s.isAllocated() && !set.isAllocated() {
		return s.Clone()
	}

	if !s.isAllocated() && set.isAllocated() {
		return set.Clone()
	}

	var resultSet Set[T]
	var setForRange Set[T]

	//copy the largest
	if s.Len() > set.Len() {
		resultSet = s.Clone()
		setForRange = set
	} else {
		resultSet = set.Clone()
		setForRange = *s
	}

	for item := range setForRange {
		if !s.Contains(item) {
			resultSet.Add(item)
		} else {
			resultSet.Remove(item)
		}
	}

	return resultSet
}

// Difference возвращает новое множество, содержащее элементы текущего множества,
// которые отсутствуют в множестве set (разность множеств: s \ set).
//
// Пример:
//
//	a := NewSet(1, 2, 3)
//	b := NewSet(2, 3, 4)
//	c := a.Difference(b) // c содержит {1}
func (s *Set[T]) Difference(set Set[T]) Set[T] {
	if !s.isAllocated() {
		return NewSet[T]()
	}

	resultSet := s.Clone()
	for item := range set {
		resultSet.Remove(item)
	}

	return resultSet
}

// Union возвращает объединение текущего множества и множества set.
// Использует стратегию Optimization (см. UnionStrategy) для выбора оптимального способа.
// Эквивалентно вызову UnionByStrategy(s, set, Optimization).
//
// Пример:
//
//	a := NewSet(1, 2)
//	b := NewSet(3, 4)
//	c := a.Union(b) // c содержит {1, 2, 3, 4}
func (s *Set[T]) Union(set Set[T]) Set[T] {
	return s.UnionByStrategy(set, Optimization)
}

// UnionByStrategy возвращает объединение множеств с использованием указанной стратегии.
// Параметр strategy определяет, какое множество будет клонировано, а какое будет итерировано.
// Доступные стратегии: Optimization (умный выбор), CurrentSet (клонировать текущее),
// IncomingSet (клонировать входящее).
//
// Пример:
//
//	a := NewSet(1, 2)
//	b := NewSet(3, 4)
//	c := a.UnionByStrategy(b, CurrentSet) // клонируется a, добавляются элементы b
func (s *Set[T]) UnionByStrategy(set Set[T], strategy UnionStrategy) Set[T] {
	switch strategy {
	case CurrentSet:
		resultSet := set.Clone()
		maps.Copy(resultSet, *s)
		return resultSet
	case IncomingSet:
		resultSet := s.Clone()
		maps.Copy(resultSet, set)
		return resultSet
	default:
		return s.unionOptimization(set)
	}
}

func (s *Set[T]) unionOptimization(set Set[T]) Set[T] {
	var resultSet Set[T]

	if s.Len() > set.Len() {
		resultSet = s.Clone()
		maps.Copy(resultSet, set)
	} else {
		resultSet = set.Clone()
		maps.Copy(resultSet, *s)
	}

	return resultSet
}

// Remove удаляет элемент item из множества.
// Если элемент отсутствует или множество nil, метод ничего не делает.
//
// Пример:
//
//	s := NewSet(1, 2, 3)
//	s.Remove(2) // s содержит {1, 3}
func (s *Set[T]) Remove(item T) {
	if !s.isAllocated() {
		return
	}

	delete(*s, item)
}

// ToSlice преобразует множество в срез (slice) типа []T.
// Порядок элементов не гарантируется (зависит от итерации по мапе).
// Для nil-множества возвращает nil.
//
// Пример:
//
//	s := NewSet("a", "b", "c")
//	slice := s.ToSlice() // []string{"a", "b", "c"} в произвольном порядке
func (s *Set[T]) ToSlice() []T {
	if !s.isAllocated() {
		return nil
	}

	items := make([]T, 0, s.Len())

	for item := range *s {
		items = append(items, item)
	}

	return items
}

// Len возвращает количество элементов в множестве.
// Для nil-множества возвращает 0.
//
// Пример:
//
//	s := NewSet(1, 2, 3)
//	s.Len() // 3
func (s *Set[T]) Len() int {
	if !s.isAllocated() {
		return 0
	}

	return len(*s)
}

// Clone возвращает глубокую копию множества.
// Для nil-множества возвращает пустое выделенное множество (не nil).
//
// Пример:
//
//	original := NewSet(1, 2, 3)
//	copy := original.Clone() // copy содержит {1, 2, 3}, но является отдельным множеством
func (s *Set[T]) Clone() Set[T] {
	if !s.isAllocated() {
		return NewSet[T]()
	}

	return maps.Clone(*s)
}

// Reset очищает множество, удаляя все элементы.
// Если множество nil, оно инициализируется как пустое выделенное множество.
//
// Пример:
//
//	s := NewSet(1, 2, 3)
//	s.Reset() // s теперь пусто
func (s *Set[T]) Reset() {
	if s == nil {
		return
	}

	if *s == nil {
		*s = NewSet[T]()
		return
	}

	clear(*s)
}

// Replace заменяет содержимое текущего множества на содержимое множества set.
// Текущее множество сначала очищается (Reset), а затем в него копируются все элементы из set.
// Полезно для обновления множества без создания нового объекта.
//
// Пример:
//
//	s := NewSet(1, 2)
//	t := NewSet(3, 4)
//	s.Replace(t) // s теперь содержит {3, 4}
func (s *Set[T]) Replace(set Set[T]) {
	s.Reset()

	if *s == nil {
		*s = NewSet[T]()
	}

	maps.Copy(*s, set)
}

// Equal проверяет, равны ли два множества (содержат одинаковые элементы).
// Два множества считаются равными, если они имеют одинаковый размер
// и каждый элемент одного множества содержится в другом.
// nil-множество считается равным пустому выделенному множеству.
//
// Пример:
//
//	a := NewSet(1, 2, 3)
//	b := NewSet(3, 2, 1)
//	c := NewSet(1, 2)
//	a.Equal(b) // true
//	a.Equal(c) // false
func (s *Set[T]) Equal(set Set[T]) bool {
	if s.Len() != set.Len() {
		return false
	}

	if !s.isAllocated() && !set.isAllocated() {
		return true
	}

	if !s.isAllocated() && set.isAllocated() {
		return false
	}

	if s.isAllocated() && !set.isAllocated() {
		return false
	}

	for item := range set {
		if !s.Contains(item) {
			return false
		}
	}

	return true
}

// Filter возвращает новое множество, содержащее только те элементы текущего множества,
// для которых функция predicate возвращает true.
// Исходное множество не изменяется.
//
// Пример:
//
//	s := NewSet(1, 2, 3, 4, 5)
//	evens := s.Filter(func(x int) bool { return x%2 == 0 }) // evens содержит {2, 4}
func (s *Set[T]) Filter(predicate func(T) bool) Set[T] {
	result := NewSet[T]()

	if !s.isAllocated() {
		return result
	}

	for item := range *s {
		if predicate(item) {
			result.Add(item)
		}
	}

	return result
}

func (s *Set[T]) isAllocated() bool {
	return s != nil && *s != nil
}

// MarshalJSON реализует интерфейс json.Marshaler.
// Множество сериализуется в JSON-массив элементов.
// Для nil-множества возвращает "null".
//
// Пример:
//
//	s := NewSet("a", "b")
//	data, _ := json.Marshal(s) // data равно ["a", "b"]
func (s *Set[T]) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	return json.Marshal(s.ToSlice())
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler.
// Десериализует JSON-массив в множество. Поддерживает null и пустые массивы.
//
// Пример:
//
//	var s Set[int]
//	json.Unmarshal([]byte("[1,2,3]"), &s) // s содержит {1, 2, 3}
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || strings.ToLower(string(data)) == "null" {
		s.Reset()
		return nil
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	*s = NewSet[T](items...)
	return nil
}

// Value реализует интерфейс driver.Valuer для использования с database/sql.
// Позволяет сохранять множество в базу данных как JSON-массив.
func (s *Set[T]) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}

	return s.MarshalJSON()
}

// Scan реализует интерфейс sql.Scanner для чтения множества из базы данных.
// Поддерживает чтение JSON-массивов из колонок типов []byte или string.
func (s *Set[T]) Scan(value any) error {
	if value == nil {
		s.Reset()
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan Set[%T]", value)
	}

	if len(data) == 0 {
		*s = NewSet[T]()
		return nil
	}

	return s.UnmarshalJSON(data)
}
