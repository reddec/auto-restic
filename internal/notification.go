package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Notification struct {
	Operation string    `json:"operation"` // backup or prune
	Started   time.Time `json:"started"`
	Finished  time.Time `json:"finished"`
	Failed    bool      `json:"failed"`
	Error     string    `json:"error,omitempty"`
}

type HTTPNotification struct {
	URL           string        `long:"url" env:"URL" description:"URL to invoke"`
	Retries       int           `long:"retries" env:"RETRIES" description:"Number of additional retries" default:"5"`
	Interval      time.Duration `long:"interval" env:"INTERVAL" description:"Interval between attempts" default:"12s"`
	Method        string        `long:"method" env:"METHOD" description:"HTTP method" default:"POST"`
	Timeout       time.Duration `long:"timeout" env:"TIMEOUT" description:"Request timeout" default:"30s"`
	Authorization string        `long:"authorization" env:"AUTHORIZATION" description:"Authorization header value"`
}

func (ht *HTTPNotification) Auto(ctx context.Context, op string, fn func() error) {
	log.Println(op, "started")
	a := time.Now()
	err := fn()
	b := time.Now()
	var errMsg string
	if err == nil {
		log.Println(op, "complete")
	} else {
		log.Println(op, "failed:", err)
		errMsg = err.Error()
	}

	if ht.URL == "" {
		return
	}

	_ = ht.Notify(ctx, &Notification{
		Operation: op,
		Started:   a,
		Finished:  b,
		Failed:    errMsg != "",
		Error:     errMsg,
	})
}

func (ht *HTTPNotification) Notify(ctx context.Context, record *Notification) error {
	left := ht.Retries
	for {
		err := ht.notify(record)
		if err == nil {
			log.Println("HTTP notification delivered")
			return nil
		}

		if left <= 0 {
			break
		}
		log.Println(left, "attempts left;", "notification failed:", err)

		left--
		select {
		case <-time.After(ht.Interval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("all attempts failed")
}

func (ht *HTTPNotification) notify(message *Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), ht.Timeout)
	defer cancel()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, ht.Method, ht.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if ht.Authorization != "" {
		req.Header.Set("Authorization", ht.Authorization)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("status: %d", res.StatusCode)
	}
	return nil
}
