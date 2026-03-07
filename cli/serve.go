package cli

import (
	"log"
	"net/http"
	"time"

	"github.com/mahavirnahata/gophant"
)

type ServeOptions struct {
	Addr string
}

func Serve(app *gophant.App, opts ServeOptions) error {
	addr := app.Config.Addr
	if opts.Addr != "" {
		addr = opts.Addr
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           app.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting %s on %s", app.Config.AppName, addr)
	return server.ListenAndServe()
}
