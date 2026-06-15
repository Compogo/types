package ranger

// inclusiveBoundValue представляет включающую границу диапазона.
// Значение границы входит в диапазон.
//
// Примеры:
//   - Нижняя включающая граница: значение >= 5
//   - Верхняя включающая граница: значение <= 10
type inclusiveBoundValue[T Numeric] struct {
	baseBoundValue[T]
}

// NewInclusiveLowerBoundValue создаёт новую включающую нижнюю границу.
// Возвращает интерфейс LowerBounder для использования в качестве нижней границы диапазона.
//
// Пример:
//
//	lower := NewInclusiveLowerBoundValue(5) // нижняя граница: значение >= 5
func NewInclusiveLowerBoundValue[T Numeric](value T) LowerBounder[T] {
	return newInclusiveBoundValue[T](value)
}

// NewInclusiveUpperBoundValue создаёт новую включающую верхнюю границу.
// Возвращает интерфейс UpperBounder для использования в качестве верхней границы диапазона.
//
// Пример:
//
//	upper := NewInclusiveUpperBoundValue(10) // верхняя граница: значение <= 10
func NewInclusiveUpperBoundValue[T Numeric](value T) UpperBounder[T] {
	return newInclusiveBoundValue[T](value)
}

// NewOpenLowerBoundValue создаёт новую открытую нижнюю границу.
// Открытая граница означает отсутствие ограничения снизу (-∞).
// Возвращает интерфейс LowerBounder.
//
// Пример:
//
//	lower := NewOpenLowerBoundValue[int]() // нижняя граница: -∞
func NewOpenLowerBoundValue[T Numeric]() LowerBounder[T] {
	boundValue := newInclusiveBoundValue[T](minValue[T]())
	boundValue.boundType = OPEN

	return boundValue
}

// NewOpenUpperBoundValue создаёт новую открытую верхнюю границу.
// Открытая граница означает отсутствие ограничения сверху (+∞).
// Возвращает интерфейс UpperBounder.
//
// Пример:
//
//	upper := NewOpenUpperBoundValue[int]() // верхняя граница: +∞
func NewOpenUpperBoundValue[T Numeric]() UpperBounder[T] {
	boundValue := newInclusiveBoundValue[T](maxValue[T]())
	boundValue.boundType = OPEN

	return boundValue
}

// newInclusiveBoundValue — внутренний конструктор для создания включающей границы.
func newInclusiveBoundValue[T Numeric](value T) *inclusiveBoundValue[T] {
	return &inclusiveBoundValue[T]{
		baseBoundValue[T]{
			value:     value,
			boundType: INCLUSIVE,
		},
	}
}

// IsAboveLowerBound проверяет, что значение находится выше или равно нижней границе.
// Для включающей границы: val >= boundValue.value
func (boundValue *inclusiveBoundValue[T]) IsAboveLowerBound(val T) bool {
	return val >= boundValue.value
}

// IsBelowUpperBound проверяет, что значение находится ниже или равно верхней границе.
// Для включающей границы: val <= boundValue.value
func (boundValue *inclusiveBoundValue[T]) IsBelowUpperBound(val T) bool {
	return val <= boundValue.value
}
