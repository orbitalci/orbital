# more code more problems

## Questions
- one repo, multiple pipelines? how to handle that 
    - for example, the docker builder pattern. when deps are updated, the docker base image shoudl be as well. that's all in one repo, but its widely different behavior than build/testing source code.



## in progress



## little TODOs?:
- create `ocelot init` command
- env variable for vault address to be used by get_ssh_key.sh from s3
- fix PR's triggering builds
- something that says X isn't tracked by ocelot (ADD THIS CHECK TO ALL COMMANDS SO THAT BEHAVIOR IS CONSISTENT) 
- something to take care of removing dead docker containers + images from werker's host (this shit builds up fast)
- add something so that containers aren't removed (easy debugging for us) 
- add ability to remove webhooks??? (this would be handy while we're playing around with stuff)
- add optional to specify working directory inside of ocelot.yml
- `ocelot summary` command takes in -repo or --hash? 
- add `who triggered this` value to summary table
- add *what* triggered this build
	- command line trigger
	- pr trigger
	- commit trigger
	- etc. 



    
## bugs: 
- our fucking build is broken
- be able to properly handle KILLS (what happens when build is killed halfway?)
    - [stackoverflow](https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in-a-defe)
- fix dev mode (maybe fuck this)
- fix html viewer displaying *oldest* matching git hash (maybe fuck this)     
- all services should check on startup if they can connect to everything they depend on, then either bail or profusely log / do retries 




## BIG TODOs:
- use the [go test tagging thing](https://stackoverflow.com/questions/24030059/skip-some-tests-with-go-test)
- polling option to add new repos with ocelot.yml
- make the client colors for everything configurable
    - this would also include making it so that the ocelot client can be configured via a config.yml 
- actually parse out exit codes, not just shit itself if it gets a non-zero one
- add a custom stage where you can build base images to run your build off of?
- put a limit on number of running containers at once
- failure notifications (integration with email or slack)
- actions to only take based on branch or w/e 
    - solution: add `trigger/skip` section to stage yml?
        - implementation: when hookhandler receives a message, it will filter on new commit messages and branch to take out stages that do not fit trigger requirements
- tag built projects by custom group name, that way you can filter to see all repos belonging to your group
- do the pipeline thing
- make it so that successful builds FROM PR'S will edit the PR comments and say whether or not that build was successful 
- check out worker queue
	- storing how long your commit waited on the queue
- return to old package install list??
    - [docker get script for ex](https://get.docker.com/)
    - this also needs to save the image to nexus and reuse    



## done:
~~- create a nice fancy markdown page explaining~~ 
    ~~- how to get started~~
    ~~- all the useful commands~~
~~- detect acct/repo using `git config --get remote.origin.url`~~
~~- docker login? - our repo creds model works for this currently, just need to implement part that actually runs docker login~~
~~- `ocelot kill <hash>` - add a quit channel~~
~~- `ocelot watch` - create a new webhook~~
~~- `ocelot build jessishank/mytestocy <hash>` - to put on queue w/o bitbucket webhook (marianne)~~
~~- `ocelot watch` - create a new webhook (marianne)~~
~~- add ability to specify if you want all branches built~~   
~~- don't create webhook unless they have an ocelot.yaml file~~
~~- set environment properties that are always avilable on a build container~~
~~- GOTTA FIGURE OUT WHAT TO DO ABOUT HANDLING SSH KEYS~~
    ~~- take in a private key via command line for an account name~~ 
~~- add nexus to infra~~
- ~~remove volume mounts on spawned build containers - they should be downloading the bash scripts out of s3~~ 
- ~~make it so that we can query whether or not there's a key uploaded for this accountname (also update help)~~
~~- [BLOCKED GO TALK TO JESSI] `ocelot status` - asks admin to get build runtime (marianne)~~ 
    ~~- current / past stage info to be added to build_stage_details (currently build_failure_reason)~~
    ~~- this should really give you success/failure of some sort~~
        - ~~if running or failed, display stages + corresponding messages~~ 
    ~~- queryable by:~~
        ~~- repoName ~~
        ~~- acctname/repoName~~
        ~~- git hash~~
- sweep through repo and add updates to db at any point of failure; some areas that i can think of off hand: (jessi)                                     
    - werker dies... should update that hash somehow with build failure reason -> dead werker (at least, possibly also a re-queue)
         - panic recovery on main function 
            - cleanup consul entry / notify _someone_ of status 
            - [1](https://blog.golang.org/defer-panic-and-recover), [2](https://golangbot.com/panic-and-recover/)
            - cleanup docker containers
            - add item back to queue for build 

 
