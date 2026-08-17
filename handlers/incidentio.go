package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	// ODFE runs an action only while a trigger condition holds and offers no
	// recovery hook, so alerts are never resolved from here.
	alertStatusFiring = "firing"

	incidentIOTimeout     = 10 * time.Second
	incidentIOMaxErrBytes = 512
)

var incidentIOClient = &http.Client{Timeout: incidentIOTimeout}

// IncidentIO used to configure common params for creating alerts
type IncidentIO struct {
	URL          string
	Token        string
	DefaultTitle string
}

// incidentIO used by the handler to set params per incoming request
type incidentIO struct {
	*IncidentIO `yaml:"-"`

	Title       string
	SourceURL   string `yaml:"source_url"`
	Metadata    map[string]interface{}
	description string
}

// alertEvent is the request payload of the incident.io HTTP alert source API
type alertEvent struct {
	Title            string `json:"title"`
	Status           string `json:"status"`
	DeduplicationKey string `json:"deduplication_key"`
	Description      string `json:"description,omitempty"`
	SourceURL        string `json:"source_url,omitempty"`

	// incident.io extracts alert attributes out of this object, using the
	// expressions configured on the alert source.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (i incidentIO) create() error {
	payload, err := json.Marshal(alertEvent{
		Title:  i.Title,
		Status: alertStatusFiring,
		// incident.io groups events by this key, so a title that is stable
		// across monitor executions collapses them into a single alert.
		DeduplicationKey: i.Title,
		Description:      i.description,
		SourceURL:        i.SourceURL,
		Metadata:         i.Metadata,
	})

	if err != nil {
		return fmt.Errorf("cannot marshal alert event, %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, i.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("cannot create request, %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+i.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := incidentIOClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Accepted events come back as 202, so 200 alone is too narrow.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, incidentIOMaxErrBytes))

		return fmt.Errorf("got %s, %s", resp.Status, bytes.TrimSpace(body))
	}

	return nil
}

// EchoHandler creates an incident.io alert per each incoming http request
func (i IncidentIO) EchoHandler(c echo.Context) error {
	if i.URL == "" || i.Token == "" {
		response := "incident.io alert was not sent, 'incident.io.url' and 'incident.io.token' cli arguments were not provided"

		log.Error(response)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, response)
	}

	alerter := incidentIO{
		IncidentIO: &i,
		Title:      i.DefaultTitle,
	}

	defer c.Request().Body.Close()
	description, err := parseBody(c.Request().Body, &alerter)

	if err != nil {
		response := fmt.Sprintf("incident.io alert was not sent, %v", err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	alerter.description = description

	if err := alerter.create(); err != nil {
		response := fmt.Sprintf("incident.io alert was not sent, title: %s, %v", alerter.Title, err)

		log.Error(response)
		return echo.NewHTTPError(http.StatusInternalServerError, response)
	}

	response := fmt.Sprintf("incident.io alert successfully sent, title: %s", alerter.Title)
	log.Info(response)

	return echo.NewHTTPError(http.StatusOK, response)
}
