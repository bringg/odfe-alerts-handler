# OpenDistro for Elasticsearch Alerts Handler

An HTTP server which used to handle webhooks triggered by [OpenDistro for Elasticsearch Alerting](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting)

## Why?

As for time  of writing `destination` options that `ODFE` provides are limited.

1. It is not possible to post to different Slack channels using same Incoming Webhook URL, see [issue](https://github.com/opendistro-for-elasticsearch/alerting-kibana-plugin/issues/85)
2. It is not possible to send emails

## Features

- Ability to handle emails, and even send emails to multiple addresses within same webhook
- Ability to post to multiple slack channels within same webhook

## Install

Download latest version for your platform from [releases](https://github.com/bringg/odfe-alerts-handler/releases) page

## With Docker

    docker run --rm bringg/ofde-alerts-handler

## Usage

    usage: odfe-alerts-handler [<flags>]

    Flags:
    -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
        --web.listen-address=":8080"
                                Address to listen on for incoming HTTP requests.
        --smtp.host="localhost"  SMTP server hostname.
        --smtp.port=25           SMTP server port.
        --smtp.username=""       SMTP server login username.
        --smtp.password=""       SMTP server login password.
        --smtp.from="opendistro@localhost"
                                SMTP from address.
        --smtp.default-subject="Opendistro Alert fired"
                                SMTP default subject.
        --slack.token=""         Slack token for posting messages.

## Setup ODFE Alerting destinations

TBA

## License

Licensed under the MIT License. See the LICENSE file for details.
