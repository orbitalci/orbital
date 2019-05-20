# 2019 Roadmap

The Level11 internal installation has crossed into 1 year of continous operation! In 2019, we're switching into a heavy maintanence mode to better prepare for features.

## Project related
### Renaming the repo to OrbitalCI
Renaming the project from Ocelot to OrbitalCI is important for growing the project's discoverability. This was a problem that had continued to compound as the repo changed owners.

### Major refactoring to codebase
Ocelot has proven itself to be successful after a year in production. However the codebase has some architectural shortcomings attributed to having stronger values for quick and experimental feedback. Now that we have a better idea of what out users want, we are going to attempt to correct leaky abstractions. Particularly where they involve database and grpc interaction.

Since these changes have potential to change expected behavior, we will be using the rename and the `orb` single-binary (more on that later) as testing grounds for beginning these corrections.

### Single binary deployments
Based on internal operation, we've concluded that building multiple binaries has made it really easy to make mistakes deploying updates.

Also, it creates unnecessary distance from potential operators by separating the client cli from the server modes. 

The new binary will be named `orb`.

The `ocelot` client and the server components `admin`, `werker`, `hookhandler`, `poller` and `changecheck` will be maintained until their functionality is offered within `orb`. Though, there is potential that some newer functionality will not land in those components.

### RFC process for large changes
As an effort to be more transparent, we're trying to adopt some processes that create documentation to be created for features prior to their code landing in master. This will let stakeholders provide feedback early. This process is inspired by the Rust-lang governance model and we'll probably cherry-pick solutions that are appropriate for our current scale.

## Feature related
### Subscribe builds
Biggest new feature is the ability to configure build dependencies. Internally this has been called "subscribe builds".

The idea is that we want a build timeline when a common repo's successful build  (such as a library) may trigger a downstream repo build to start based on the downstream repo's explicit subscription to the library.