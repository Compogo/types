package mapper

import (
	"errors"
	"fmt"
	"iter"
	"reflect"
	"strings"

	typeName "github.com/Compogo/tools/type_name"
)

var (
	DoesNotExistError = errors.New("does not exist")
)

// Mapper provides a bidirectional mapping between string keys and typed values.
// It is particularly useful for enums, configuration keys, and any scenario where
// you need to convert between string representations and concrete types.
//
// The type requires T to implement fmt.Stringer, and all keys are stored in lowercase
// to ensure case-insensitive lookups.
//
// Basic usage:
//
//	mapper := NewMapper(StatusPending, StatusActive, StatusClosed)
//	status, err := mapper.Get("active")      // returns StatusActive
//	exists := mapper.HasByKey("pending")     // true
//	exists := mapper.HasByValue(StatusClosed) // true
//
// The zero value is not ready for use and must be created via NewMapper.
// All methods are safe for concurrent read-only access, but concurrent writes
// (Add) require external synchronization.
type Mapper[T fmt.Stringer] struct {
	items     map[string]T
	typeName  string
	zeroValue T
}

func NewMapper[T fmt.Stringer](values ...T) *Mapper[T] {
	mapper := &Mapper[T]{
		items:     make(map[string]T),
		typeName:  fmt.Sprintf("Mapper[%s]", typeName.TypeName[T]()),
		zeroValue: *new(T),
	}

	mapper.Add(values...)

	return mapper
}

func (mapper *Mapper[T]) Add(values ...T) {
	for _, value := range values {
		mapper.items[strings.ToLower(value.String())] = value
	}
}

func (mapper *Mapper[T]) Get(key string) (T, error) {
	if val, exists := mapper.items[strings.ToLower(key)]; exists {
		return val, nil
	}

	return mapper.zeroValue, fmt.Errorf("key %s %w for type %s", key, DoesNotExistError, mapper.typeName)
}

func (mapper *Mapper[T]) Keys() []string {
	keys := make([]string, 0, len(mapper.items))
	for k := range mapper.items {
		keys = append(keys, k)
	}

	return keys
}

func (mapper *Mapper[T]) HasByKey(key string) bool {
	_, exists := mapper.items[strings.ToLower(key)]
	return exists
}

func (mapper *Mapper[T]) HasByValue(value T) bool {
	for _, v := range mapper.items {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}

	return false
}

func (mapper *Mapper[T]) RemoveByValue(value T) {
	mapper.RemoveByKey(value.String())
}

func (mapper *Mapper[T]) RemoveByKey(key string) {
	delete(mapper.items, strings.ToLower(key))
}

func (mapper *Mapper[T]) All() iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for key, value := range mapper.items {
			if !yield(key, value) {
				return
			}
		}
	}
}
