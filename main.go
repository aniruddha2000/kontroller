package main

import (
	"github.com/aniruddha2000/kontroller/api"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/cli/globalflag"
	"net/http"
	"os"
	"time"
)

func main() {
	webhookServer := api.NewWebhookServer()

	fs := pflag.NewFlagSet(api.Kon, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, api.Kon)
	webhookServer.Opt.AddFlagSet(fs)

	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(webhookServer.Handler.KlusterValidationHandler))

	webhookServer.Cfg = webhookServer.Opt.Config()

	stopCh := server.SetupSignalHandler()
	ch, _, err := webhookServer.Cfg.SecInfo.Serve(mux, 10*time.Second, stopCh)
	if err != nil {
		panic(err)
	}
	<-ch
}
