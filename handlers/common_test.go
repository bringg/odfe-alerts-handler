package handlers

import (
	"io"
	"strings"
	"testing"
)

func newBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func TestParseBodyParsesSlackParams(t *testing.T) {
	slacker := slack{Slack: &Slack{}}

	data, err := parseBody(newBody("channels: ['#alerts']\nusers: ['test@example.com']\n---\nthe body\n"), &slacker)
	if err != nil {
		t.Fatalf("parseBody returned error: %v", err)
	}

	if len(slacker.Channels) != 1 || slacker.Channels[0] != "#alerts" {
		t.Errorf("channels = %v, want [#alerts]", slacker.Channels)
	}

	if len(slacker.Users) != 1 || slacker.Users[0] != "test@example.com" {
		t.Errorf("users = %v, want [test@example.com]", slacker.Users)
	}

	if data != "the body\n" {
		t.Errorf("data = %q, want %q", data, "the body\n")
	}
}

func TestParseBodyParsesEmailParams(t *testing.T) {
	emailer := email{Email: &Email{}}

	data, err := parseBody(newBody("to: ['test@example.com']\nsubject: a subject\n---\nthe body\n"), &emailer)
	if err != nil {
		t.Fatalf("parseBody returned error: %v", err)
	}

	if len(emailer.To) != 1 || emailer.To[0] != "test@example.com" {
		t.Errorf("to = %v, want [test@example.com]", emailer.To)
	}

	if emailer.Subject != "a subject" {
		t.Errorf("subject = %q, want %q", emailer.Subject, "a subject")
	}

	if data != "the body\n" {
		t.Errorf("data = %q, want %q", data, "the body\n")
	}
}
