package linker

import (
	"fmt"
	"iter"

	typeName "github.com/Compogo/tools/type_name"
	"github.com/Compogo/types/mapper"
)

type Linker[T comparable, I any] struct {
	values    map[T]I
	typeName  string
	zeroValue I

	normalizer Normalizer[T]
}

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

func (linker *Linker[T, I]) Add(key T, value I) {
	linker.values[linker.normalizer(key)] = value
}

func (linker *Linker[T, I]) Get(key T) (I, error) {
	if val, exists := linker.values[linker.normalizer(key)]; exists {
		return val, nil
	}

	return linker.zeroValue, fmt.Errorf("key %s %w for type %s", key, mapper.DoesNotExistError, linker.typeName)
}

func (linker *Linker[T, I]) GetOrDefault(key T, d I) I {
	if val, exists := linker.values[linker.normalizer(key)]; exists {
		return val
	}

	return d
}

func (linker *Linker[T, I]) Has(key T) bool {
	_, exists := linker.values[linker.normalizer(key)]
	return exists
}

func (linker *Linker[T, I]) Remove(key T) {
	delete(linker.values, linker.normalizer(key))
}

func (linker *Linker[T, I]) Len() int {
	return len(linker.values)
}

func (linker *Linker[T, I]) Keys() []T {
	keys := make([]T, 0, len(linker.values))
	for k := range linker.values {
		keys = append(keys, k)
	}

	return keys
}

func (linker *Linker[T, I]) Reset() {
	clear(linker.values)
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
