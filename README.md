# OpenDistro for Elasticsearch Alerts Handler

An HTTP server which used to handle webhooks triggered by [OpenDistro for Elasticsearch Alerting](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting)

> Notice, the readme is for `0.3.x` version

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

```shell
docker run --rm -p 8080:8080 bringg/odfe-alerts-handler --help
```

## Usage

```plain
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
```

## Configure ODFE Alerting destinations

### First

1. Go to `Alerting` > `Destinations`
2. Create the destination with type `Custom webhook`
3. Choose `Define endpoint by URL`
    - For `slack` set the url to have path with `/slack`, like `http://odfe-server:8080/slack`
    - For `email` set the url to have path with `/email`, like `http://odfe-server:8080/email`

### Sending Email from triggers

1. Select destination which was created with the `/email` path
2. The `Message` body look like below:

```yaml
to: ['example@test.com']
subject: Optional subject param
---
This is the body of the message
Here you can use the templeting as usual...
```

`subject` is optional, if not provided the default one used, see [usage](#usage).

### Sending Slack from triggers

1. Select destination which was created with the `/slack` path
2. The `Message` body look like below:

```yaml
channels: ['#alerts']
users: ['test@example.com']
---
This is the body of the message
Here you can use the templeting as usual...
```

You can have both `channels` and `users` keys if you desire to send to both.
Optionally, for `channels` you can omit the leading `#`.

## Creating a release

```shell
RELEASE_TITLE="Module maintenance, and fix slack deprecations"
RELEASE_VERSION=0.3.1

git tag -a v${RELEASE_VERSION} -m "${RELEASE_TITLE}"
git push --tags
goreleaser --rm-dist
```

## License

Licensed under the MIT License. See the LICENSE file for details.
