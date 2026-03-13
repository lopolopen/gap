package gapc

import (
	"github.com/lopolopen/gap/options/gap"
	"github.com/lopolopen/shoot"
)

var (
	GoGenerated               = gap.GoGenerated
	HandleTopicWithinGroupRaw = gap.HandleTopicWithinGroupRaw
	FileName                  = gap.FileName
	FuncName                  = gap.FuncName
	Resolve                   = gap.Resolve
)

func ResolveType[T any](typeName string, resolve func(v T)) shoot.Option[gap.HandlerDepsOptions, *gap.HandlerDepsOptions] {
	return gap.ResolveType(typeName, resolve)
}
