package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	emailClient "github.com/jordan-wright/email"
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

// HTTPHandler handles incoming http request
// 1: Get the data from the request body
// 2: Prepare email for sending (optionally get subject from the body)
// 3: Call email.send method
func (e Email) HTTPHandler(resp http.ResponseWriter, req *http.Request) {
	addresses := strings.Split(mux.Vars(req)["address"], ",")
	response := fmt.Sprintf("email successfuly sent, to: %v\n", addresses)

	defer req.Body.Close()
	requestBody, err := ioutil.ReadAll(req.Body)

	if err != nil {
		response = fmt.Sprintf("email was not sent, failed to read body, %v", err)

		log.Print(response)
		http.Error(resp, response, http.StatusInternalServerError)
		return
	}

	emailer := email{
		Email: e,
		data:  requestBody,
		to:    addresses,
	}

	emailer.prepareSubject()

	if err := emailer.send(); err != nil {
		response = fmt.Sprintf("email was not sent, to: %v, subject: %s, %v", emailer.to, emailer.subject, err)

		log.Print(response)
		http.Error(resp, response, http.StatusInternalServerError)
		return
	}

	log.Print(response)
	fmt.Fprint(resp, response)
}
