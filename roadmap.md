# 2020 Roadmap

## Stable release target for 2020
The main goal for this year is to get to a stable release published to both Github and Crates.io. We'll break the path to this target down into domains.

The project has no sponsors at the moment, so progress will be slowing down until further notice. The focus will be closing/fixing issues filed from the Go codebase. An alpha release should be expected to publish early 2020.

### Backend
* Complete the rest of the service endpoints, including changes for what it will take to stream live logs to the client in chunked form
* Polling repos to support auto-start builds on new commits
* MacOS Docker support
* Non-containerized builds
* Slack notifications
* Repo subscription build triggers

### CLI
* Streaming live logs
* CLI client file-based configs
* Server data import-export

## Suspending contribution structure for adoption
The project hasn't attracted any developers. In order to not discourage anyone who potentially wants to contribute, plain PRs will suffice. I'll continue to try my best to keep communication of design and project direction open.