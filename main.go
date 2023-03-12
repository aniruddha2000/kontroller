package main

import (
	"net/http"
	"os"
	"time"

	"github.com/aniruddha2000/kontroller/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/cli/globalflag"
)

func main() {
	webhookServer := api.NewWebhookServer()

	start := time.Now()
	log.Infof("Starting @ %s", start.String())

	log.SetLevel(log.DebugLevel)

	fs := pflag.NewFlagSet(api.Kon, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, api.Kon)
	webhookServer.Opt.AddFlagSet(fs)

	if err := fs.Parse(os.Args); err != nil {
		log.Errorf("flag parse: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(webhookServer.Handler.KlusterValidationHandler))

	webhookServer.Cfg = webhookServer.Opt.Config()

	stopCh := server.SetupSignalHandler()
	ch, _, err := webhookServer.Cfg.SecInfo.Serve(mux, 10*time.Second, stopCh)
	if err != nil {
		log.Errorf("Error serving webhook: %v", err)
	}
	<-ch
}
