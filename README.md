# OpenDistro for Elasticsearch Alerts Handler

An HTTP server which used to handle webhooks triggered by [OpenDistro for Elasticsearch Alerting](https://opendistro.github.io/for-elasticsearch-docs/docs/alerting)

> Notice, the readme is for `0.4.x` version

## Why?

As for time  of writing `destination` options that `ODFE` provides are limited.

1. It is not possible to post to different Slack channels using same Incoming Webhook URL, see [issue](https://github.com/opendistro-for-elasticsearch/alerting-kibana-plugin/issues/85)
2. It is not possible to send emails
3. It is not possible to build a structured JSON body, which alerting APIs such as [incident.io](https://incident.io) require.

## Features

- Ability to handle emails, and even send emails to multiple addresses within same webhook
- Ability to post to multiple slack channels and/or users within same webhook
- Ability to create [incident.io](https://incident.io) alerts, for driving escalations

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
    --incident.io.url=""     incident.io HTTP alert source URL.
    --incident.io.token=""   incident.io alert source token.
    --incident.io.default-title="Opendistro Alert fired"
                            incident.io default alert title.
```

## Configure ODFE Alerting destinations

### First

1. Go to `Alerting` > `Destinations`
2. Create the destination with type `Custom webhook`
3. Choose `Define endpoint by URL`
    - For `slack` set the url to have path with `/slack`, like `http://odfe-server:8080/slack`
    - For `email` set the url to have path with `/email`, like `http://odfe-server:8080/email`
    - For `incident.io` set the url to have path with `/incident.io`, like `http://odfe-server:8080/incident.io`

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

### Sending incident.io alerts from triggers

First create an [HTTP alert source](https://docs.incident.io/alerts/alert-sources) in incident.io, then start the
handler with the URL and token it gives you:

```shell
odfe-alerts-handler \
  --incident.io.url="https://api.incident.io/v2/alert_events/http/<ALERT_SOURCE_CONFIG_ID>" \
  --incident.io.token="<ALERT_SOURCE_TOKEN>"
```

1. Select destination which was created with the `/incident.io` path
2. The `Message` body look like below:

```yaml
title: High error rate on checkout
source_url: https://kibana.example.com/app/alerting
metadata:
  severity: critical
  cluster:
    name: prod
    region: us-east-1
---
This is the body of the message, it becomes the alert description.
Here you can use the templeting as usual...
```

| Param | Required | Description |
| --- | --- | --- |
| `title` | no | Defaults to `--incident.io.default-title`, and doubles as the deduplication key |
| `source_url` | no | Link back to the alert origin, shown in incident.io |
| `metadata` | no | Values for incident.io to extract alert attributes from, nesting allowed |

`metadata` is passed through to the alert event untouched, where nested objects and lists are both fine. It
only takes effect once the alert source is configured to extract from it, under `Attributes` on the source in
incident.io, where each attribute binds to an ES5 expression over the payload:

```javascript
$.metadata.severity
$.metadata.cluster.region
```

Until then it is carried along but unused.

The `title` is sent as the deduplication key. ODFE runs the trigger action on *every* monitor execution while
the condition holds, and incident.io groups events by that key, so a title which is stable across executions
collapses them into a single alert:

```yaml
# good, stable across executions
title: "High error rate on {{ctx.monitor.name}}"

# bad, a new alert every single run
title: "{{ctx.results.0.hits.total.value}} errors on checkout"
```

Give every trigger its own title. Anything left on the default shares a deduplication key with every other
such trigger, and they all collapse into one alert.

Alerts are always sent as `firing`.

## Creating a release

```shell
RELEASE_TITLE="Ability to create incident.io alerts, for driving escalations"
RELEASE_VERSION=0.4.0

git tag -a v${RELEASE_VERSION} -m "${RELEASE_TITLE}"
git push --tags
goreleaser --rm-dist
```

## License

Licensed under the MIT License. See the LICENSE file for details.
