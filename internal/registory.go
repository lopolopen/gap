package internal

import (
	"fmt"
	"reflect"

	"github.com/lopolopen/gap/internal/enum"
)

var extRegistries = make(map[enum.PluginType]any)

func Register[T any](typ enum.PluginType, value T) {
	extRegistries[typ] = value
}

func MustGet[T any](typ enum.PluginType) T {
	var zero T
	if typ == enum.None {
		return zero
	}
	v, _ := extRegistries[typ]
	t, ok := v.(T)
	if !ok {
		elem := reflect.TypeOf((*T)(nil)).Elem()
		panic(fmt.Sprintf("key '%s': got %T, want %s", typ, v, elem.String()))
	}
	return t
}
