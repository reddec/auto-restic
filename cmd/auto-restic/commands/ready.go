package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Ready struct {
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"API call timeout" default:"3s"`
	Address string        `long:"api-address" env:"API_ADDRESS" description:"Address for internal API" default:"127.0.0.1:8080"`
}

func (cmd *Ready) Execute([]string) error {
	global, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	url := "http://" + cmd.Address + "/ready"

	ctx, timed := context.WithTimeout(global, cmd.Timeout)
	defer timed()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}

	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("%d %s", res.StatusCode, res.Status)
	}

	return nil
}
