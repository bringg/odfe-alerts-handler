package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"

	emailClient "github.com/jordan-wright/email"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

var subjectRe = regexp.MustCompile(`(?i)^(\s+)?subject:(\s+)?`)

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
	Email

	subject string
	to      []string
	data    []byte
}

func (e *email) prepareSubject() {
	scanner := bufio.NewScanner(bytes.NewReader(e.data))
	scanner.Scan()

	if subjectRe.Match(scanner.Bytes()) {
		e.subject = subjectRe.ReplaceAllString(scanner.Text(), "")
		e.data = e.data[len(scanner.Bytes()):]
		return
	}

	e.subject = e.DefaultSubject
}

func (e *email) send() error {
	client := &emailClient.Email{
		To:      e.to,
		From:    e.From,
		Subject: e.subject,
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
	var addresses []string

	if c.QueryParam("addresses") != "" {
		addresses = strings.Split(c.QueryParam("addresses"), ",")
	}

	if len(addresses) == 0 {
		response := "email was not sent, no addresses param provided"

		log.Error(response)
		return echo.NewHTTPError(http.StatusBadRequest, response)
	}

	defer c.Request().Body.Close()
	requestBody, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		response := fmt.Sprintf("email was not sent, failed to read body, %v", err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	emailer := email{
		Email: e,
		data:  requestBody,
		to:    addresses,
	}

	emailer.prepareSubject()

	if err := emailer.send(); err != nil {
		response := fmt.Sprintf("email was not sent, to: %v, subject: %s, %v", addresses, emailer.subject, err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	response := fmt.Sprintf("email successfuly sent, to: %v, subject: %s", addresses, emailer.subject)
	log.Info(response)
	return echo.NewHTTPError(http.StatusOK, response)
}
