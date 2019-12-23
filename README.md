# anyslk

`* -> Slack message`

## Support protocol

### :email: SMTP

Mail -> Slack message

``` console
$ export SLACK_INCOMMING_WEBHOOK_URL=https://hooks.slack.com/services/XXXXXXXXX/XXXXXXXXX/XXXxxxXXXXXX
$ anyslk --listen-smtp --smtp-port=1025
```

#### Usage

Convert `RCPT TO` address to Slack channel name

```
random@example.com -> Post message to #random channel
^^^^^^                                 ^^^^^^
```

### :earth_asia: HTTP

:construction:
