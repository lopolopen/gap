package gap

import (
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/options/gap"
	"github.com/lopolopen/gap/storage"
	"github.com/lopolopen/shoot"
)

type groupedSubs struct {
	subMap      map[string]*internal.Sub
	handlerOtps []gap.HandlerOptions
	diOtps      []gap.DIOptions
}

var subs *groupedSubs

func Subscribe(opts ...shoot.Option[Options, *Options]) {
	gapOpts := new(Options).With(opts...)

	ds := gapOpts.Dependencies()
	if subs == nil {
		subs = &groupedSubs{}
	}
	subs.diOtps = append(subs.diOtps, ds...)
	hs := gapOpts.Handlers()
	subs.handlerOtps = append(subs.handlerOtps, hs...)

	if gap.RegisterHandlerOnly(gapOpts) {
		return
	}

	initSnowflake(gapOpts.WorkerID)
	initDashboard(gapOpts)

	for _, dep := range subs.diOtps {
		dep.Resolve(gapOpts.Values())
	}

	err := subs.subscribe(gapOpts)
	if err != nil {
		panic(err)
	}

	err = subs.listeningAll()
	if err != nil {
		panic(err)
	}
}

func (g *groupedSubs) subscribe(gapOpts *Options) error {
	for _, o := range g.handlerOtps {
		if o.Handler == nil {
			return errx.ErrNilHandler
		}
		if o.Topic == "" {
			return errx.ErrEmptyTopic
		}
		opt := *gapOpts
		group := o.Group
		if group == "" {
			group = opt.DefaultGroup
		}
		opt.Group = group

		if g.subMap == nil {
			g.subMap = make(map[string]*internal.Sub)
		}
		sub, ok := g.subMap[group]
		if !ok {
			stor := internal.MustGet[storage.Storage](opt.StorageExt())
			var brok broker.Broker
			sub = internal.NewSub(&opt, stor, brok)
			g.subMap[group] = sub
		}
		err := sub.Subscribe(o.Topic, o.Handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *groupedSubs) listeningAll() error {
	for _, sub := range g.subMap {
		err := sub.Listening()
		if err != nil {
			return err
		}
	}
	return nil
}
