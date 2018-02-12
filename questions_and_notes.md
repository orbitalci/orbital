# more code more problems

## Questions
- handling ssh keys or just  `go` & `dep ensure` problems 
- one repo, multiple pipelines? how to handle that 
    - for example, the docker builder pattern. when deps are updated, the docker base image shoudl be as well. that's all in one repo, but its widely different behavior than build/testing source code.

## in progress
- nexus (jessi)
- change so that we store info about every stage (marianne)
    - `ocelot status` - asks admin to get build runtime 
        - current / past stage info to be added to build_stage_details (currently build_failure_reason)
        - queryable by:
            - acctname 
            - acctname/repoName
            - git hash

## bugs: 
- run test
- fix dev mode (maybe fuck this)
- fix html viewer displaying *oldest* matching git hash (maybe fuck this)


## BIG TODOs:
- actually parse out exit codes, not just shit itself if it gets a non-zero one
- failure notifications
- actions to only take based on branch or w/e 
    - possible solution: add `trigger` section to stage yml?
        - implementation: when hookhandler receives a message, it will filter on new commit messages and branch to take out stages that do not fit trigger requirements
- sweep through repo and add updates to db at any point of failure; some areas that i can think of off hand:
    - failed validation at hookhandler stage 
    - failed setup stage
    - more verbose for other stage failures
    - werker dies... should update that hash somehow with build failure reason -> dead werker (at least, possibly also a re-queue)
         - panic recovery on main function 
            - cleanup consul entry / notify _someone_ of status 
            - [RECOVERY!!!](https://blog.golang.org/defer-panic-and-recover)
- docker login? - our repo creds model works for this currently, need to implement
- tighter maven integration?
- `ocelot kill <hash>` - add a quit channel
- check out worker queue
	- storing how long your commit waited on the queue 
- return to old package install list??
    - [docker get script for ex](https://get.docker.com/)    
    
## little TODOs?: 
- `ocelot trigger jessishank/mytestocy <hash>` - to put on queue w/o bitbucket webhook         
- polling option? idk
- `ocelot summary` comand takes in --acct or --hash ????
- would be cool if we could take in regex like `--acct-repo=level11consulting/orchestr8*` ?  
- add `who triggered this` value to summary table
- add *what* triggered this build
	- command line trigger
	- pr trigger
	- commit trigger
	- etc. 

## done: 
~~- when printing matching git hashes, should display corresponding acctname/repo like `ocelot summary` command~~
~~- change client's --validate command to take in a value just like all the other commands~~
~~- hash matching only works when you pass in full hash~~
