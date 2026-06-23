package ranger

import "math"

// Numeric определяет интерфейс всех числовых типов, которые могут быть использованы
// в диапазонах. Включает целые (знаковые и беззнаковые) и числа с плавающей точкой.
//
// Поддерживаемые типы:
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// minValue возвращает минимальное значение для числового типа T.
// Для целых чисел возвращает наименьшее представимое значение (например, MinInt64).
// Для чисел с плавающей точкой возвращает -MaxFloat (приближение к -∞).
// Для беззнаковых целых возвращает 0.
func minValue[T Numeric]() T {
	var minValue T
	switch any(minValue).(type) {
	case int:
		return any(math.MinInt).(T)
	case int8:
		return any(math.MinInt8).(T)
	case int16:
		return any(math.MinInt16).(T)
	case int32:
		return any(math.MinInt32).(T)
	case int64:
		return any(math.MinInt64).(T)
	case float32:
		return any(-float32(math.MaxFloat32)).(T)
	case float64:
		return any(-math.MaxFloat64).(T)
	}

	return T(0)
}

// maxValue возвращает максимальное значение для числового типа T.
// Для целых и беззнаковых чисел возвращает наибольшее представимое значение.
// Для чисел с плавающей точкой возвращает MaxFloat (приближение к +∞).
func maxValue[T Numeric]() T {
	var minValue T
	switch any(minValue).(type) {
	case int:
		return any(math.MaxInt).(T)
	case int8:
		return any(math.MaxInt8).(T)
	case int16:
		return any(math.MaxInt16).(T)
	case int32:
		return any(math.MaxInt32).(T)
	case int64:
		return any(math.MaxInt64).(T)
	case uint:
		return any(math.MaxUint32).(T)
	case uint8:
		return any(math.MaxUint8).(T)
	case uint16:
		return any(math.MaxUint16).(T)
	case uint32:
		return any(math.MaxUint32).(T)
	case uint64:
		return any(math.MaxUint64).(T)
	case float32:
		return any(float32(math.MaxFloat32)).(T)
	case float64:
		return any(math.MaxFloat64).(T)
	}

	return T(0)
}
