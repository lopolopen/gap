package internal

import (
	"fmt"

	"github.com/lopolopen/gap/internal/enum"
)

var extRegistries = make(map[enum.ExtType]any)

func Register[T any](typ enum.ExtType, value T) {
	extRegistries[typ] = value
}

func MustGet[T any](typ enum.ExtType) T {
	var zero T
	if typ == enum.None {
		return zero
	}
	v, ok := extRegistries[typ]
	if !ok {
		panic(fmt.Sprintf("no value is registered with the given type: %s", typ))
	}
	t, ok := v.(T)
	if !ok {
		var x T
		panic(fmt.Sprintf("value registered by type %s is not type: %T", typ, x))
	}
	return t
}
