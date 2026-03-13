package gap

import "github.com/lopolopen/gap/internal/enum"

type PluginOptions interface {
	PluginType() enum.PluginType
}
