package set

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

const numberOfElementsToRecreate = 5

// Set is a generic collection type that stores unique elements of type T.
//
// It can be used as a replacement for map[T]struct{} with a cleaner API.
// The zero value (nil) is ready to use and represents an empty set.
//
// The type implements:
//   - Basic operations: Add, Remove, Contains, Len, ToSlice
//   - Set operations: Append (union), Clone
//   - JSON marshaling/unmarshaling
//   - SQL driver.Valuer and sql.Scanner interfaces
//
// All methods are safe to call on nil receiver and will not panic.
type Set[T comparable] map[T]struct{}

var exists = struct{}{}

func NewSet[T comparable](items ...T) Set[T] {
	s := Set[T]{}

	s.Add(items...)

	return s
}

func (s *Set[T]) Add(items ...T) {
	if s == nil || len(items) == 0 {
		return
	}

	if len(items) >= numberOfElementsToRecreate {
		newSet := make(map[T]struct{}, s.Len()+uint32(len(items)))

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

func (s *Set[T]) Append(set Set[T]) {
	if s == nil || set.Len() == 0 {
		return
	}

	newSet := make(map[T]struct{}, s.Len()+set.Len())

	if *s != nil {
		for item := range *s {
			newSet[item] = exists
		}
	}

	for item := range set {
		newSet[item] = exists
	}

	*s = newSet
}

func (s *Set[T]) Contains(item T) bool {
	if s == nil || *s == nil {
		return false
	}

	_, isExists := (*s)[item]
	return isExists
}

func (s *Set[T]) Remove(item T) {
	if s == nil || *s == nil {
		return
	}

	delete(*s, item)
}

func (s *Set[T]) ToSlice() []T {
	if s == nil || *s == nil {
		return nil
	}

	items := make([]T, 0, s.Len())

	for item := range *s {
		items = append(items, item)
	}

	return items
}

func (s *Set[T]) Len() uint32 {
	if s == nil || *s == nil {
		return 0
	}

	return uint32(len(*s))
}

func (s *Set[T]) Clone() Set[T] {
	if s == nil || *s == nil {
		return NewSet[T]()
	}

	return maps.Clone(*s)
}

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

func (s *Set[T]) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	return json.Marshal(s.ToSlice())
}

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

func (s *Set[T]) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}

	return s.MarshalJSON()
}

func (s *Set[T]) Scan(value interface{}) error {
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
