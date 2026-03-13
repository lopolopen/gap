package gap

import (
	"github.com/lopolopen/gap/broker"
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/storage"
)

func mustGetBroker(gapOpts *Options) broker.Broker {
	bp := gapOpts.BrokerPlugin
	if bp == nil {
		return nil
	}
	f := internal.MustGet[broker.Factory](bp.PluginType())
	brok, err := f.CreateBroker(gapOpts)
	if err != nil {
		panic(err)
	}
	return brok
}

func mustGetStorage(gapOpts *Options) storage.Storage {
	sp := gapOpts.StoragePlugin
	if sp == nil {
		return nil
	}
	f := internal.MustGet[storage.Factory](sp.PluginType())
	stor, err := f.CreateStorage(gapOpts)
	if err != nil {
		panic(err)
	}
	return stor
}
