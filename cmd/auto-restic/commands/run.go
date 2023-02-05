package commands

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/reddec/auto-restic/internal"
)

type Run struct {
	Address string          `long:"api-address" env:"API_ADDRESS" description:"Address for internal API" default:"127.0.0.1:8080"`
	Backup  internal.Config `group:"Backup config" namespace:"backup" env-namespace:"BACKUP"`
}

func (cmd *Run) Execute([]string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	app := internal.New(cmd.Backup)

	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
		if app.Ready() {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusUnprocessableEntity)
		}
	})
	srv := http.Server{Addr: cmd.Address, Handler: mux}
	defer srv.Close()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return app.Run(ctx)
}
