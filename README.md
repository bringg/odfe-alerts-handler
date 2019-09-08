# OpenDistro for Elasticsearch Alerts Handler

An HTTP server which used to handle webhooks triggered by [OpenDistro for Elasticsearch Alerting](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting)

## Why?

As for time  of writing `destination` options that `ODFE` provides are limited.

1. It is not possible to post to different Slack channels using same Incoming Webhook URL, see [issue](https://github.com/opendistro-for-elasticsearch/alerting-kibana-plugin/issues/85)
2. It is not possible to send emails

## Features

- Ability to handle emails, and even send emails to multiple addresses within same webhook
- Ability to post to multiple slack channels and/or users within same webhook

## Install

Download latest version for your platform from [releases](https://github.com/bringg/odfe-alerts-handler/releases) page

## With Docker

    docker run --rm -p 8080:8080 bringg/odfe-alerts-handler --help

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

## Configure ODFE Alerting destinations

### First

1. Go to `Alerting` > `Destinations`
2. Create the destination with type `Custom webhook`
3. Chose `Define endpoint by custom attributes URL`

Fill in `Type`, `Host` and `Port` according to how and where you installed `odfe-alerts-handler`.

### Configuring for Email

1. Set `Path` to `email`
2. Set `Query parameters` as follows:
    - Key: `addresses`
    - Value: comma separated list of emails to send

You can also override the default subject.
If the first line of the alert message contains `Subject:`, that line will be used as a subject for the email.

### Configuring for Slack

1. Set `Path` to `slack`
2. Set `Query parameters` as follows:
    - Key: `channels` or `users`
    - Value: comma separated list of user **emails** or list of channels

You can have both `channels` and `users` keys if you desire to send to both.
Optionally, for `channels` you can omit the leading `#`.

## Creating a release

```shell
RELEASE_TITLE="First release"
RELEASE_VERSION=0.1.0

git tag -a v${RELEASE_VERSION} -m "${RELEASE_TITLE}"
git push --tags
goreleaser --rm-dist
```

## License

Licensed under the MIT License. See the LICENSE file for details.
