package handlers

import (
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strconv"

	emailClient "github.com/jordan-wright/email"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// Email used to configure common params for sending email
type Email struct {
	Host           string
	Port           int
	Username       string
	Password       string
	From           string
	DefaultSubject string
}

// email used by the handler to set params per incoming request
type email struct {
	*Email

	Subject string
	To      []string
	data    []byte
}

func (e *email) send() error {
	client := &emailClient.Email{
		To:      e.To,
		From:    e.From,
		Subject: e.Subject,
		Text:    e.data,
	}

	var auth smtp.Auth
	if e.Username != "" && e.Password != "" {
		auth = smtp.PlainAuth("", e.Username, e.Password, e.Host)
	}

	return client.Send(net.JoinHostPort(e.Host, strconv.Itoa(e.Port)), auth)
}

// EchoHandler sends email per each incoming http request
func (e Email) EchoHandler(c echo.Context) error {
	emailer := email{
		Email:   &e,
		Subject: e.DefaultSubject,
	}

	defer c.Request().Body.Close()
	data, err := parseBody(c.Request().Body, &emailer)

	if err != nil {
		response := fmt.Sprintf("email was not sent, %v", err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	if len(emailer.To) == 0 {
		response := "email was not sent, 'To' param wasn't provided"

		log.Error(response)
		return echo.NewHTTPError(http.StatusBadRequest, response)
	}

	emailer.data = []byte(data)

	if err := emailer.send(); err != nil {
		response := fmt.Sprintf("email was not sent, to: %v, subject: %s, %v", emailer.To, emailer.Subject, err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	response := fmt.Sprintf("email successfuly sent, to: %v, subject: %s", emailer.To, emailer.Subject)
	log.Info(response)
	return echo.NewHTTPError(http.StatusOK, response)
}
