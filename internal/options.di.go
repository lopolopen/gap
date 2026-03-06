package internal

import (
	"fmt"

	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot new -opt -short -type=DIOptions

type DIOptions struct {
	FileName   string
	PkgPath    string
	FuncName   string
	_resolvers []func(values []any) string
}

func ResolveType[T any](typeName string, resolve func(v T)) shoot.Option[DIOptions, *DIOptions] {
	return func(o *DIOptions) {
		o._resolvers = append(o._resolvers, func(values []any) string {
			var dep T
			var resolved bool
			for _, v := range values {
				if d, ok := v.(T); ok {
					dep = d
					resolved = true
					break
				}
			}
			if !resolved {
				return typeName
			}
			resolve(dep)
			return ""
		})
	}
}

func Resolve(opts ...shoot.Option[DIOptions, *DIOptions]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(DIOptions).With(opts...)
		o._dependencies = append(o._dependencies, *options)
	}
}

func (d *DIOptions) Resolve(values []any) {
	var unresolved string
	for _, r := range d._resolvers {
		tname := r(values)
		if tname != "" {
			unresolved += fmt.Sprintf("\n🧩 %s", tname)
		}
	}
	if len(unresolved) > 0 {
		panic(fmt.Sprintf("unresolved dependencies at %s: %s", d.FuncName, unresolved))
	}
}
