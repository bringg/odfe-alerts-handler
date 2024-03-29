package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	slackAPI "github.com/slack-go/slack"
)

// Slack used to configure common params for posting messages
type Slack struct {
	Token  string
	client *slackAPI.Client
}

// slack used by the handler to set params per incoming request
type slack struct {
	*Slack `yaml:"-"`

	Channels []string
	Users    []string
	text     string
}

func (s *slack) getClient() *slackAPI.Client {
	if s.client == nil {
		s.client = slackAPI.New(s.Token)
	}

	return s.client
}

func (s slack) postToChannels() error {
	var result error

	for _, channel := range s.Channels {
		if !strings.HasPrefix(channel, "#") {
			channel = "#" + channel
		}

		if _, _, err := s.getClient().PostMessage(channel, slackAPI.MsgOptionText(string(s.text), false)); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (s slack) postToUsers() error {
	var result error

	for _, user := range s.Users {

		user, err := s.getClient().GetUserByEmail(user)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_, _, err = s.getClient().PostMessage(user.ID, slackAPI.MsgOptionText(string(s.text), false), slackAPI.MsgOptionAsUser(true))
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (s slack) post() error {
	var result error

	if len(s.Channels) > 0 {
		if err := s.postToChannels(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if len(s.Users) > 0 {
		if err := s.postToUsers(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

// EchoHandler posts slack message per each incoming http request
func (s Slack) EchoHandler(c echo.Context) error {
	if s.Token == "" {
		response := "slack message was not sent, 'slack.token' cli argument was not provided"

		log.Error(response)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, response)
	}

	slacker := slack{
		Slack: &s,
	}

	defer c.Request().Body.Close()
	text, err := parseBody(c.Request().Body, &slacker)

	if err != nil {
		response := fmt.Sprintf("slack message was not sent, %v", err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	if len(slacker.Channels) == 0 && len(slacker.Users) == 0 {
		response := "slack message was not sent, no channels or users params provided"

		log.Error(response)
		return echo.NewHTTPError(http.StatusBadRequest, response)
	}

	slacker.text = text

	if err := slacker.post(); err != nil {
		response := fmt.Sprintf("slack message was not sent, channels: %v, users: %v, %v", slacker.Channels, slacker.Users, err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	response := fmt.Sprintf("slack message successfuly sent, channels: %v, users: %v", slacker.Channels, slacker.Users)
	log.Info(response)
	return echo.NewHTTPError(http.StatusOK, response)
}
