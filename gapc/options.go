package gapc

import (
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/shoot"
)

var (
	GoGenerated               = internal.GoGenerated
	HandleTopicWithinGroupRaw = internal.HandleTopicWithinGroupRaw
	FileName                  = internal.FileName
	FuncName                  = internal.FuncName
	Resolve                   = internal.Resolve
)

func ResolveType[T any](typeName string, resolve func(v T)) shoot.Option[internal.DIOptions, *internal.DIOptions] {
	return internal.ResolveType(typeName, resolve)
}
