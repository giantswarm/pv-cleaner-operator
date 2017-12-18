package flag

import (
	"github.com/giantswarm/microkit/flag"

	"github.com/giantswarm/endpoint-operator/flag/service"
)

type Flag struct {
	Service service.Service
}

func New() *Flag {
	f := &Flag{}
	flag.Init(f)

	return f
}
