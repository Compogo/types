// Package hash_slice предоставляет структуру данных HashSlice,
// сочетающую возможности слайса и хэш-таблицы.
//
// Особенности:
//   - Сохраняет порядок добавления элементов
//   - O(1) доступ по индексу и поиск по значению
//   - Автоматическая гарантия уникальности элементов
//   - Удобные методы для удаления по значению или индексу
//   - Поддержка итерации через range
//
// Сравнение с другими структурами:
//   | Операция          | Slice | Map | HashSlice |
//   |-------------------|-------|-----|-----------|
//   | Доступ по индексу | O(1)  | -   | O(1)      |
//   | Поиск по значению | O(n)  | O(1)| O(1)      |
//   | Уникальность      | ручная| да  | да        |
//   | Сохранение порядка| да    | нет | да        |
//
// Пример использования:
//
//	// Создание
//	hs := NewHashSlice[string]()
//	hs.Add("first")
//	hs.Add("second")
//
//	// Поиск
//	idx := hs.IndexOf("first") // 0
//	ok := hs.Contains("second") // true
//
//	// Доступ
//	val, _ := hs.GetByIndex(1) // "second"
//
//	// Удаление
//	hs.Remove("first")
//
//	// Итерация
//	for i, v := range hs.All() {
//	    fmt.Printf("[%d] %v\n", i, v)
//	}

package hash_slice

import (
	"errors"
	"fmt"
	"iter"

	typeName "github.com/Compogo/tools/type_name"
	"github.com/Compogo/types/linker"
)

// NoneIndex — специальное значение, возвращаемое методом IndexOf,
// если элемент не найден в HashSlice.
const NoneIndex = -1

// Ошибки, возвращаемые при работе с HashSlice.
var (
	// AlreadyExistsError возникает при попытке добавить уже существующий элемент.
	// HashSlice гарантирует уникальность элементов, повторное добавление запрещено.
	AlreadyExistsError = errors.New("already exists")

	// OutOfRangeError возникает при попытке доступа к элементу по индексу,
	// выходящему за пределы допустимого диапазона.
	OutOfRangeError = errors.New("out of range")
)

// HashSlice представляет собой структуру данных, сочетающую слайс и хэш-таблицу.
// Сохраняет порядок добавления элементов, обеспечивает быстрый поиск по значению
// и гарантирует уникальность всех элементов.
//
// Внутренняя реализация:
//   - Слайс items хранит элементы в порядке добавления
//   - Linker хранит отображение элемент -> индекс для быстрого поиска
//
// Тип T должен быть сравниваемым (comparable).
type HashSlice[T comparable] struct {
	linker    *linker.Linker[T, int]
	items     []T
	typeName  string
	zeroValue T
}

// NewHashSliceFrom создаёт новый HashSlice и заполняет его переданными элементами.
// Если среди переданных элементов есть дубликаты, возвращает ошибку AlreadyExistsError.
//
// Пример:
//
//	hs, err := NewHashSliceFrom("apple", "banana", "cherry")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	hs, err := NewHashSliceFrom(1, 2, 2) // ошибка: элемент 2 уже существует
func NewHashSliceFrom[T comparable](items ...T) (*HashSlice[T], error) {
	hs := NewHashSlice[T]()
	hs.items = make([]T, 0, len(items))

	if err := hs.Append(items...); err != nil {
		return nil, err
	}

	return hs, nil
}

// NewHashSlice создаёт пустой HashSlice.
//
// Пример:
//
//	hs := NewHashSlice[string]()
//	hs.Add("hello")
//	hs.Add("world")
func NewHashSlice[T comparable]() *HashSlice[T] {
	return &HashSlice[T]{
		linker:    linker.NewLinker[T, int](),
		items:     nil,
		typeName:  fmt.Sprintf("HashSlice[%s]", typeName.TypeName[T]()),
		zeroValue: *new(T),
	}
}

// Len возвращает количество элементов в HashSlice.
// Сложность: O(1)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("a", "b", "c")
//	fmt.Println(hs.Len()) // 3
func (hs *HashSlice[T]) Len() int {
	return len(hs.items)
}

// Append добавляет несколько элементов в конец HashSlice.
// Если какой-либо из элементов уже существует, возвращает ошибку AlreadyExistsError.
//
// Пример:
//
//	hs := NewHashSlice[string]()
//	err := hs.Append("apple", "banana", "cherry")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (hs *HashSlice[T]) Append(items ...T) (err error) {
	for _, item := range items {
		if _, err = hs.Add(item); err != nil {
			return err
		}
	}

	return nil
}

// Add добавляет один элемент в конец HashSlice.
// Возвращает индекс добавленного элемента.
// Если элемент уже существует, возвращает ошибку AlreadyExistsError.
//
// Пример:
//
//	hs := NewHashSlice[string]()
//	idx, err := hs.Add("apple")
//	fmt.Println(idx) // 0
//
//	_, err = hs.Add("apple") // ошибка: already exists
func (hs *HashSlice[T]) Add(item T) (int, error) {
	if err := hs.checkContains(item); err != nil {
		return NoneIndex, err
	}

	hs.items = append(hs.items, item)
	index := len(hs.items) - 1

	hs.linker.Add(item, index)

	return index, nil
}

// GetByIndex возвращает элемент по индексу.
// Если индекс выходит за пределы диапазона, возвращает ошибку OutOfRangeError.
// Сложность: O(1)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	item, err := hs.GetByIndex(1)
//	fmt.Println(item) // banana
//
//	_, err = hs.GetByIndex(10) // ошибка: out of range
func (hs *HashSlice[T]) GetByIndex(index int) (T, error) {
	if err := hs.checkIndex(index); err != nil {
		return hs.zeroValue, err
	}

	return hs.items[index], nil
}

// Items возвращает срез всех элементов в порядке добавления.
// Возвращаемый срез является ссылкой на внутренние данные — не изменяйте его!
// Сложность: O(1)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("a", "b", "c")
//	for _, item := range hs.Items() {
//	    fmt.Println(item)
//	}
func (hs *HashSlice[T]) Items() []T {
	return hs.items
}

// IndexOf возвращает индекс элемента в HashSlice.
// Если элемент не найден, возвращает NoneIndex (-1).
// Сложность: O(1)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	fmt.Println(hs.IndexOf("banana")) // 1
//	fmt.Println(hs.IndexOf("grape"))  // -1
func (hs *HashSlice[T]) IndexOf(item T) int {
	if !hs.linker.Has(item) {
		return NoneIndex
	}

	index, _ := hs.linker.Get(item)
	return index
}

// Remove удаляет элемент из HashSlice по его значению.
// Если элемент не найден, метод ничего не делает.
// После удаления индексы всех последующих элементов уменьшаются на 1.
// Сложность: O(n) из-за перестроения индексов
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	hs.Remove("banana")
//	// теперь hs содержит: ["apple", "cherry"]
//	fmt.Println(hs.IndexOf("cherry")) // 1 (было 2)
func (hs *HashSlice[T]) Remove(item T) {
	index := hs.IndexOf(item)
	if index != NoneIndex {
		_ = hs.RemoveByIndex(index)
	}
}

// RemoveByIndex удаляет элемент по индексу.
// Возвращает ошибку OutOfRangeError, если индекс невалиден.
// После удаления индексы всех последующих элементов уменьшаются на 1.
// Сложность: O(n) из-за перестроения индексов
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	err := hs.RemoveByIndex(1) // удаляет "banana"
//	// теперь hs содержит: ["apple", "cherry"]
func (hs *HashSlice[T]) RemoveByIndex(index int) (err error) {
	if err = hs.checkIndex(index); err != nil {
		return err
	}

	hs.items = append(hs.items[:index], hs.items[index+1:]...)

	hs.linker.Reset()
	for i, item := range hs.items {
		hs.linker.Add(item, i)
	}

	return nil
}

// Contains проверяет, содержится ли элемент в HashSlice.
// Сложность: O(1)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana")
//	fmt.Println(hs.Contains("apple"))  // true
//	fmt.Println(hs.Contains("grape"))  // false
func (hs *HashSlice[T]) Contains(item T) bool {
	return hs.linker.Has(item)
}

// Replace заменяет элемент по указанному индексу на новый.
// Возвращает ошибку, если индекс невалиден или новый элемент уже существует.
// Сложность: O(1)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	err := hs.Replace(1, "grape") // заменяет "banana" на "grape"
//	// теперь hs содержит: ["apple", "grape", "cherry"]
//
//	err = hs.Replace(1, "apple") // ошибка: already exists
func (hs *HashSlice[T]) Replace(index int, item T) (err error) {
	if err = hs.checkIndex(index); err != nil {
		return err
	}

	if err = hs.checkContains(item); err != nil {
		return err
	}

	hs.items[index] = item
	hs.linker.Add(item, index)

	return nil
}

// Reset очищает HashSlice, удаляя все элементы.
// Сложность: O(n)
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	hs.Reset()
//	fmt.Println(hs.Len()) // 0
func (hs *HashSlice[T]) Reset() {
	hs.linker.Reset()
	clear(hs.items)
}

// All возвращает итератор для обхода всех элементов HashSlice.
// Реализует интерфейс iter.Seq2, позволяя использовать range в цикле.
// Возвращает пары (индекс, значение) в порядке добавления элементов.
//
// Пример:
//
//	hs, _ := NewHashSliceFrom("apple", "banana", "cherry")
//	for index, value := range hs.All() {
//	    fmt.Printf("%d: %s\n", index, value)
//	}
//	// Вывод:
//	// 0: apple
//	// 1: banana
//	// 2: cherry
//
// Итерацию можно прервать досрочно:
//
//	for index, value := range hs.All() {
//	    if value == "banana" {
//	        break
//	    }
//	}
func (hs *HashSlice[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, item := range hs.items {
			if !yield(i, item) {
				return
			}
		}
	}
}

// checkIndex проверяет валидность индекса.
// Возвращает OutOfRangeError, если индекс выходит за пределы диапазона.
func (hs *HashSlice[T]) checkIndex(index int) error {
	if index < 0 || index >= hs.Len() {
		return fmt.Errorf("%s index %d %w", hs.typeName, index, OutOfRangeError)
	}
	return nil
}

// checkContains проверяет, не существует ли уже элемент в HashSlice.
// Возвращает AlreadyExistsError, если элемент уже присутствует.
func (hs *HashSlice[T]) checkContains(item T) error {
	if hs.Contains(item) {
		return fmt.Errorf("%s item %v %w", hs.typeName, item, AlreadyExistsError)
	}
	return nil
}
