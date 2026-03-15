package linker

import (
	"fmt"
	"iter"

	typeName "github.com/Compogo/tools/type_name"
	"github.com/Compogo/types/mapper"
)

type Link[T comparable, I any] struct {
	Key   T
	Value I
}

func NewLink[T comparable, I any](key T, value I) *Link[T, I] {
	return &Link[T, I]{Key: key, Value: value}
}

// Linker stores typed values (I) indexed by keys of type T.
// It combines a Mapper for case-insensitive string lookups with
// a direct key→value map, allowing access both by string name
// and by typed key.
//
// This is useful for plugin systems, strategy patterns, or any
// scenario where you need to select implementations by name
// while maintaining type safety.
//
// Example:
//
//	//go:generate stringer -type=Driver
//
//	type Driver uint8
//	const (
//	    Postgres Driver = 0
//	    MySQL    Driver = 1
//	)
//
//	linker := NewLinker(
//	    NewLink(Postgres, &PGDriver{}),
//	    NewLink(MySQL, &MySQLDriver{}),
//	)
//	driver, _ := linker.GetByName("postgres") // returns &PGDriver{}
type Linker[T comparable, I any] struct {
	values    map[T]I
	typeName  string
	zeroValue I
}

func NewLinker[T comparable, I any](links ...*Link[T, I]) *Linker[T, I] {
	enum := &Linker[T, I]{
		values:    make(map[T]I),
		typeName:  fmt.Sprintf("Linker[%s, %s]", typeName.TypeName[T](), typeName.TypeName[I]()),
		zeroValue: *new(I),
	}

	for _, impl := range links {
		enum.Add(impl.Key, impl.Value)
	}

	return enum
}

func (linker *Linker[T, I]) Add(key T, value I) {
	linker.values[key] = value
}

func (linker *Linker[T, I]) Get(key T) (I, error) {
	if val, exists := linker.values[key]; exists {
		return val, nil
	}

	return linker.zeroValue, fmt.Errorf("key %s %w for type %s", key, mapper.DoesNotExistError, linker.typeName)
}

func (linker *Linker[T, I]) GetOrDefault(key T, d I) I {
	if val, exists := linker.values[key]; exists {
		return val
	}

	return d
}

func (linker *Linker[T, I]) Has(key T) bool {
	_, exists := linker.values[key]
	return exists
}

func (linker *Linker[T, I]) Remove(key T) {
	delete(linker.values, key)
}

func (linker *Linker[T, I]) All() iter.Seq2[T, I] {
	return func(yield func(T, I) bool) {
		for key, value := range linker.values {
			if !yield(key, value) {
				return
			}
		}
	}
}
