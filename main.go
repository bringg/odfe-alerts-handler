package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bringg/odfe-alerts-handler/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
)

const shutdownTimeout time.Duration = 60 * time.Second

func main() {
	hostname, _ := os.Hostname()

	var (
		listenAddress      = kingpin.Flag("web.listen-address", "Address to listen on for incoming HTTP requests.").Default(":8080").String()
		smtpHost           = kingpin.Flag("smtp.host", "SMTP server hostname.").Default("localhost").String()
		smtpPort           = kingpin.Flag("smtp.port", "SMTP server port.").Default("25").Int()
		smtpUsername       = kingpin.Flag("smtp.username", "SMTP server login username.").Default("").String()
		smtpPassword       = kingpin.Flag("smtp.password", "SMTP server login password.").Default("").String()
		smtpFrom           = kingpin.Flag("smtp.from", "SMTP from address.").Default(fmt.Sprintf("opendistro@%s", hostname)).String()
		smtpDefaultSubject = kingpin.Flag("smtp.default-subject", "SMTP default subject.").Default("Opendistro Alert fired").String()
		slackToken         = kingpin.Flag("slack.token", "Slack token for posting messages.").Default("").String()
	)

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	emailHandler := handlers.Email{
		Host:           *smtpHost,
		Port:           *smtpPort,
		Username:       *smtpUsername,
		Password:       *smtpPassword,
		From:           *smtpFrom,
		DefaultSubject: *smtpDefaultSubject,
	}

	slackHandler := handlers.Slack{
		Token: *slackToken,
	}

	r := mux.NewRouter()
	r.HandleFunc("/email/{address}", emailHandler.HTTPHandler)
	r.HandleFunc("/slack/{channel}", slackHandler.HTTPHandler)

	srv := &http.Server{
		Addr: *listenAddress,

		// Good practice to set timeouts :)
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		log.Printf("Starting listening on %s...", *listenAddress)
		log.Fatal(srv.ListenAndServe())
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c // Block until the signals

	// Create a deadline to wait for.
	log.Printf("Shutting down with graceful timeout of %v", shutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	srv.Shutdown(ctx)
	os.Exit(0)
}
