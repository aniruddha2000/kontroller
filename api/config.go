package api

import (
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
)

const (
	// Kon is kontroller global name
	Kon = "kontroller"
)

// Config defines the server info for custom webhook server.
type Config struct {
	SecInfo *server.SecureServingInfo
}

// Options define the server option for custom webhook server.
type Options struct {
	SecOpts *options.SecureServingOptions
}

// NewDefaultOptions return server info.
func NewDefaultOptions() *Options {
	o := &Options{
		SecOpts: options.NewSecureServingOptions(),
	}
	o.SecOpts.BindPort = 8443
	o.SecOpts.ServerCert.PairName = Kon

	return o
}

// AddFlagSet add the flags supported by default kubernetes API server.
func (o *Options) AddFlagSet(fs *pflag.FlagSet) {
	o.SecOpts.AddFlags(fs)
}

// Config return server config for custom webhook server.
func (o *Options) Config() *Config {
	err := o.SecOpts.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, nil)
	if err != nil {
		panic(err)
	}

	c := &Config{}
	err = o.SecOpts.ApplyTo(&c.SecInfo)
	if err != nil {
		panic(err)
	}

	return c
}
