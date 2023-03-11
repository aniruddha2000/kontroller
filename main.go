package main

import (
	"github.com/aniruddha2000/kontroller/api"
	"github.com/aniruddha2000/kontroller/api/handlers"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/cli/globalflag"
	"net/http"
	"os"
	"time"
)

func main() {
	option := api.NewDefaultOptions()

	fs := pflag.NewFlagSet(api.Kon, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, api.Kon)
	option.AddFlagSet(fs)

	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handlers.KlusterValidationHandler))

	stopCh := server.SetupSignalHandler()

	info := option.Config()
	ch, _, err := info.SecInfo.Serve(mux, 10*time.Second, stopCh)
	if err != nil {
		panic(err)
	}
	<-ch
}
