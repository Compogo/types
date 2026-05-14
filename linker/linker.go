package linker

import (
	"fmt"
	"iter"
	"maps"

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

func (linker *Linker[T, I]) Clone() *Linker[T, I] {
	return &Linker[T, I]{
		values:     maps.Clone(linker.values),
		typeName:   linker.typeName,
		zeroValue:  linker.zeroValue,
		normalizer: linker.normalizer,
	}
}

func (linker *Linker[T, I]) HasAllAnd(items ...T) bool {
	for _, item := range items {
		if !linker.Has(item) {
			return false
		}
	}

	return true
}

func (linker *Linker[T, I]) HasAllOr(items ...T) bool {
	for _, item := range items {
		if linker.Has(item) {
			return true
		}
	}

	return false
}

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

func (linker *Linker[T, I]) Difference(l *Linker[T, I]) *Linker[T, I] {
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

func (linker *Linker[T, I]) Union(l *Linker[T, I]) *Linker[T, I] {
	return linker.UnionByStrategy(l, Optimization)
}

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

func (linker *Linker[T, I]) Replace(l *Linker[T, I]) {
	linker.Reset()
	maps.Copy(linker.values, l.values)
}

func Copy[T comparable, I any](dst *Linker[T, I], src *Linker[T, I]) {
	for key, val := range src.values {
		dst.Add(key, val)
	}
}
