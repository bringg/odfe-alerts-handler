package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	slackAPI "github.com/nlopes/slack"
)

// Slack used to configure common params for posting messages
type Slack struct {
	Token  string
	client *slackAPI.Client
}

// slack used by the handler to set params per incoming request
type slack struct {
	Slack

	channel []string
	text    string
}

func (s *slack) send() error {
	var errors error

	if s.client == nil {
		s.client = slackAPI.New(s.Token)
	}

	for _, channel := range s.channel {

		if _, _, err := s.client.PostMessage(channel, slackAPI.MsgOptionText(string(s.text), false)); err != nil {
			multierror.Append(errors, err)
		}
	}

	return errors
}

// HTTPHandler handles incoming http request
// 1: Check if token was provider, bail out if not
// 2: Get the data from the request body
// 3: Call slack.send method
func (s Slack) HTTPHandler(resp http.ResponseWriter, req *http.Request) {
	channels := strings.Split(mux.Vars(req)["channel"], ",")
	response := fmt.Sprintf("slack message successfuly sent, channel: %v\n", channels)

	if s.Token == "" {
		response = "slack message was not sent, token was not provided"

		log.Print(response)
		http.Error(resp, response, http.StatusUnprocessableEntity)
		return
	}

	defer req.Body.Close()
	requestBody, err := ioutil.ReadAll(req.Body)

	if err != nil {
		response = fmt.Sprintf("slack message was not sent, failed to read body, %v", err)

		log.Print(response)
		http.Error(resp, response, http.StatusInternalServerError)
		return
	}

	slacker := slack{
		Slack:   s,
		channel: channels,
		text:    string(requestBody),
	}

	if err := slacker.send(); err != nil {
		response = fmt.Sprintf("slack message was not sent, channel: %v, %v", slacker.channel, err)

		log.Print(response)
		http.Error(resp, response, http.StatusInternalServerError)
		return
	}

	log.Print(response)
	fmt.Fprint(resp, response)
}
