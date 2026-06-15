// Package ranger предоставляет типобезопасную реализацию числовых диапазонов.
//
// Пакет позволяет создавать диапазоны с различными типами границ:
//   - Включающие (INCLUSIVE) — значение границы входит в диапазон
//   - Исключающие (EXCLUSIVE) — значение границы не входит в диапазон
//   - Открытые (OPEN) — отсутствие ограничения (бесконечность)
//
// Основные возможности:
//   - Создание диапазонов с валидацией границ
//   - Проверка вхождения значения в диапазон
//   - Типобезопасное разделение нижних и верхних границ на уровне компиляции
//
// Пример использования:
//
//	// Создание диапазона [0, 100) — от 0 включительно до 100 исключительно
//	lower := ranger.NewInclusiveLowerBoundValue(0)
//	upper := ranger.NewExclusiveUpperBoundValue(100)
//	r, err := ranger.NewRange(lower, upper)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Проверка значений
//	r.Contains(50)   // true
//	r.Contains(0)    // true  (включительно)
//	r.Contains(100)  // false (исключительно)
//	r.Contains(-10)  // false
//
//	// Диапазон с открытой границей: (-∞, 10]
//	openLower := ranger.NewOpenLowerBoundValue[int]()
//	inclusiveUpper := ranger.NewInclusiveUpperBoundValue(10)
//	r2, _ := ranger.NewRange(openLower, inclusiveUpper)
//	r2.Contains(-1000) // true
//	r2.Contains(10)    // true
//	r2.Contains(100)   // false

package ranger

// Ranger определяет контракт для работы с диапазоном значений.
// Позволяет получить границы диапазона и проверить вхождение значения.
type Ranger[T Numeric] interface {
	// GetLowerBound возвращает нижнюю границу диапазона.
	GetLowerBound() LowerBounder[T]
	// GetUpperBound возвращает верхнюю границу диапазона.
	GetUpperBound() UpperBounder[T]
	// Contains проверяет, входит ли значение в диапазон.
	Contains(T) bool
}

// Range представляет собой диапазон значений числового типа T.
// Содержит нижнюю и верхнюю границы, каждая из которых может быть
// включающей (INCLUSIVE), исключающей (EXCLUSIVE) или открытой (OPEN).
//
// Примеры диапазонов:
//   - [1, 10] — от 1 до 10 включительно
//   - (0, 100) — от 0 до 100 исключительно
//   - (-∞, 5] — все значения меньше или равные 5
//   - [0, +∞) — все значения больше или равные 0
type Range[T Numeric] struct {
	lower LowerBounder[T]
	upper UpperBounder[T]
}

// NewRange создаёт новый диапазон с указанными нижней и верхней границами.
// Выполняет валидацию границ:
//   - Верхняя граница должна быть больше нижней
//   - Обе границы не могут быть одновременно открытыми (OPEN)
//
// Возвращает:
//   - *Range[T] — созданный диапазон
//   - error — ошибка валидации (UpperBelowLowerError или OpenBoundError)
//
// Пример:
//
//	lower := NewInclusiveLowerBoundValue(0)
//	upper := NewExclusiveUpperBoundValue(100)
//	r, err := NewRange(lower, upper) // диапазон [0, 100)
func NewRange[T Numeric](lower LowerBounder[T], upper UpperBounder[T]) (*Range[T], error) {
	if upper.Value() <= lower.Value() {
		return nil, UpperBelowLowerError
	}

	if lower.Type() == OPEN && upper.Type() == OPEN {
		return nil, OpenBoundError
	}

	return &Range[T]{lower: lower, upper: upper}, nil
}

// GetLowerBound возвращает нижнюю границу диапазона.
// Возвращаемое значение реализует интерфейс LowerBounder.
func (ranger *Range[T]) GetLowerBound() LowerBounder[T] {
	return ranger.lower
}

// GetUpperBound возвращает верхнюю границу диапазона.
// Возвращаемое значение реализует интерфейс UpperBounder.
func (ranger *Range[T]) GetUpperBound() UpperBounder[T] {
	return ranger.upper
}

// Contains проверяет, входит ли значение value в диапазон.
// Возвращает true, если значение удовлетворяет условиям обеих границ.
//
// Пример:
//
//	r, _ := NewRange(NewInclusiveLowerBoundValue(0), NewInclusiveUpperBoundValue(10))
//	r.Contains(5)  // true
//	r.Contains(10) // true
//	r.Contains(-1) // false
func (ranger *Range[T]) Contains(value T) bool {
	return ranger.lower.IsAboveLowerBound(value) && ranger.upper.IsBelowUpperBound(value)
}
