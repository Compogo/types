package linker

import "strings"

type Option[T comparable, I any] func(l *Linker[T, I])

type Normalizer[T comparable] func(key T) T

func normalizerDefault[T comparable](key T) T {
	return key
}

func Link[T comparable, I any](key T, val I) Option[T, I] {
	return func(l *Linker[T, I]) {
		l.Add(key, val)
	}
}

func WithNormalizer[T comparable, I any](normalizer Normalizer[T]) Option[T, I] {
	return func(l *Linker[T, I]) {
		l.normalizer = normalizer
	}
}

func KeyStringNormalizer[T string, I any]() Option[T, I] {
	return WithNormalizer[T, I](func(key T) T {
		return T(strings.ToLower(string(key)))
	})
}
