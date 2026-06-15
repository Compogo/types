package linker

import "strings"

// Option представляет функцию настройки (functional option) для Linker.
// Используется для конфигурирования поведения Linker при создании.
type Option[T comparable, I any] func(l *Linker[T, I])

// Normalizer определяет функцию нормализации ключей.
// Принимает ключ и возвращает его нормализованную версию.
// По умолчанию используется функция, возвращающая ключ без изменений.
type Normalizer[T comparable] func(key T) T

// normalizerDefault — стандартная функция нормализации, возвращающая ключ без изменений.
func normalizerDefault[T comparable](key T) T {
	return key
}

// Link создаёт опцию для добавления пары (ключ, значение) в Linker.
// Полезно для начальной инициализации Linker при вызове NewLinker.
//
// Пример:
//
//	linker := NewLinker[string, int](
//	    Link("one", 1),
//	    Link("two", 2),
//	)
func Link[T comparable, I any](key T, val I) Option[T, I] {
	return func(l *Linker[T, I]) {
		l.Add(key, val)
	}
}

// WithNormalizer устанавливает пользовательскую функцию нормализации ключей.
// Нормализатор применяется ко всем операциям с ключами (Add, Get, Has, Remove).
//
// Пример:
//
//	linker := NewLinker[string, int](
//	    WithNormalizer(strings.ToUpper),
//	)
func WithNormalizer[T comparable, I any](normalizer Normalizer[T]) Option[T, I] {
	return func(l *Linker[T, I]) {
		l.normalizer = normalizer
	}
}

// KeyStringNormalizer возвращает опцию для регистронезависимой нормализации
// строковых ключей. Применяется только для Linker с типом ключа string.
// Все ключи будут приведены к нижнему регистру.
//
// Пример:
//
//	linker := NewLinker[string, int](
//	    KeyStringNormalizer[int](),
//	)
//	linker.Add("KEY", 42)
//	linker.Has("key") // true
func KeyStringNormalizer[I any]() Option[string, I] {
	return func(l *Linker[string, I]) {
		l.normalizer = keyStringNormalizer
	}
}

// keyStringNormalizer — нормализатор для строковых ключей,
// приводящий их к нижнему регистру.
func keyStringNormalizer(key string) string {
	return strings.ToLower(key)
}
