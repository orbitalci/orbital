# Notifications 

Notifications are configured using the `notify` block in the `ocelot.yml` file. Currently only slack notifications are supported. You can see useage in the [useage doc](useage.md).

```yaml
image: alpine:latest
buildTool: go
notify:
  slack:
    channel: "@jessishank"
    identifier: "ocelot-base"
    on:
      - "PASS"
      - "FAIL"
branches:
  - ALL
env:
  - "OCELOT_PATH=src/github.com/shankj3/ocelot"
stages:
  - name: build
    script:
      - cd $GOPATH/$OCELOT_PATH
      - scripts/build-release-server.sh

``` 

## Slack 
To set up a new slack webhook, first create an [incoming webhook integration](https://my.slack.com/services/new/incoming-webhook/). 

You can then add the credential to ocelot: 
```bash
ocelot creds notify add \
    --url https://hooks.slack.com/services/generated/webhook/url \
    --acctname my_team_acct --identifier my_team_id 
```

To use them in your build, add the `notify` block to your `ocelot.yml`. Pick a channel to post the notification to, and **be sure** to 
set the `identifier` tag to the value you set when adding the slack credential.   

For example, a notify block that references the cred addition above could be:
```yaml
notify:
  slack:
    channel: "#buildstatuses"
    identifier: "my_team_id"
    on:
      - "PASS"
      - "FAIL"
```