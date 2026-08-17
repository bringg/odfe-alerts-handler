package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func postToIncidentIO(t *testing.T, handler IncidentIO, body string) *echo.HTTPError {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/incident.io", strings.NewReader(body))
	c := echo.New().NewContext(req, httptest.NewRecorder())

	err := handler.EchoHandler(c)

	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("EchoHandler returned %T (%v), want *echo.HTTPError", err, err)
	}

	return he
}

// alertEventRecorder stands in for the incident.io alert source endpoint.
func alertEventRecorder(t *testing.T, payload *map[string]interface{}, header *http.Header) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*header = r.Header.Clone()

		if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
			t.Errorf("cannot decode request payload: %v", err)
		}

		// incident.io acknowledges alert events rather than returning 200.
		w.WriteHeader(http.StatusAccepted)
	}))
}

func TestIncidentIOEchoHandlerPostsAlertEvent(t *testing.T) {
	var (
		payload map[string]interface{}
		header  http.Header
	)

	srv := alertEventRecorder(t, &payload, &header)
	defer srv.Close()

	handler := IncidentIO{URL: srv.URL, Token: "a-token", DefaultTitle: "a default title"}

	he := postToIncidentIO(t, handler, "title: High error rate\n"+
		"source_url: https://kibana/app/alerting\n"+
		"metadata:\n  severity: critical\n  cluster:\n    name: prod\n"+
		"---\nthe description\n")

	if he.Code != http.StatusOK {
		t.Fatalf("code = %d (%v), want %d", he.Code, he.Message, http.StatusOK)
	}

	if got := header.Get("Authorization"); got != "Bearer a-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer a-token")
	}

	if got := header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	want := map[string]interface{}{
		"title":             "High error rate",
		"status":            "firing",
		"deduplication_key": "High error rate",
		"description":       "the description\n",
		"source_url":        "https://kibana/app/alerting",
		"metadata": map[string]interface{}{
			"severity": "critical",
			"cluster":  map[string]interface{}{"name": "prod"},
		},
	}

	if !reflect.DeepEqual(payload, want) {
		t.Errorf("payload = %#v, want %#v", payload, want)
	}
}

func TestIncidentIOEchoHandlerFallsBackToDefaultTitle(t *testing.T) {
	var (
		payload map[string]interface{}
		header  http.Header
	)

	srv := alertEventRecorder(t, &payload, &header)
	defer srv.Close()

	handler := IncidentIO{URL: srv.URL, Token: "a-token", DefaultTitle: "a default title"}

	he := postToIncidentIO(t, handler, "source_url: https://kibana/app/alerting\n---\nthe description\n")

	if he.Code != http.StatusOK {
		t.Fatalf("code = %d (%v), want %d", he.Code, he.Message, http.StatusOK)
	}

	if payload["title"] != "a default title" {
		t.Errorf("title = %v, want %q", payload["title"], "a default title")
	}

	if payload["deduplication_key"] != "a default title" {
		t.Errorf("deduplication_key = %v, want %q", payload["deduplication_key"], "a default title")
	}

	if _, ok := payload["metadata"]; ok {
		t.Errorf("metadata = %v, want it absent when no metadata is given", payload["metadata"])
	}
}

func TestIncidentIOEchoHandlerRejectsUnconfiguredHandler(t *testing.T) {
	he := postToIncidentIO(t, IncidentIO{}, "title: High error rate\n---\nthe description\n")

	if he.Code != http.StatusUnprocessableEntity {
		t.Errorf("code = %d (%v), want %d", he.Code, he.Message, http.StatusUnprocessableEntity)
	}
}
