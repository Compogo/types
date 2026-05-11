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

func KeyStringNormalizer[I any]() Option[string, I] {
	return WithNormalizer[string, I](func(key string) string {
		return strings.ToLower(key)
	})
}
