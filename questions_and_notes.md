# more code more problems

## Questions
- one repo, multiple pipelines? how to handle that 
    - for example, the docker builder pattern. when deps are updated, the docker base image shoudl be as well. that's all in one repo, but its widely different behavior than build/testing source code.

## in progress
- nexus (jessi)
- `ocelot status` - asks admin to get build runtime (marianne) 
    ~~- current / past stage info to be added to build_stage_details (currently build_failure_reason)~~
    - queryable by:
        - acctname 
        - acctname/repoName
        ~~- git hash~~
    
## bugs: 
- GOTTA FIGURE OUT WHAT TO DO ABOUT HANDLING SSH KEYS
- be able to properly handle KILLS (what happens when build is killed halfway?)
    - [stackoverflow](https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in-a-defe)
- fix dev mode (maybe fuck this)
- fix html viewer displaying *oldest* matching git hash (maybe fuck this)     

## BIG TODOs:
- actually parse out exit codes, not just shit itself if it gets a non-zero one
- failure notifications
- actions to only take based on branch or w/e 
    - possible solution: add `trigger` section to stage yml?
        - implementation: when hookhandler receives a message, it will filter on new commit messages and branch to take out stages that do not fit trigger requirements
- tag built projects by custom group name, that way you can filter to see all repos belonging to your group
- sweep through repo and add updates to db at any point of failure; some areas that i can think of off hand:
    - failed validation at hookhandler stage 
    - failed setup stage
        - should probably start printing out the actual docker errors...
    - more verbose for other stage failures
    - werker dies... should update that hash somehow with build failure reason -> dead werker (at least, possibly also a re-queue)
         - panic recovery on main function 
            - cleanup consul entry / notify _someone_ of status 
            - [RECOVERY!!!](https://blog.golang.org/defer-panic-and-recover)
- docker login? - our repo creds model works for this currently, just need to implement part that actually runs docker login
- do the pipeline thing
- tighter maven integration?
- `ocelot kill <hash>` - add a quit channel
- check out worker queue
	- storing how long your commit waited on the queue 
- return to old package install list??
    - [docker get script for ex](https://get.docker.com/)    

    
## little TODOs?: 
- add ability to specify if you want all branches built
- fix `ocelot logs ` retrieval by build_id 
- assume that the git = whatever you last pushed up (run `git rev-parse` command) 
- add ability to query for logs by build id `ocelot logs --build-id 3`
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
~~- fix goddamn tests~~ 
~~- store stages to db~~
~~- when printing matching git hashes, should display corresponding acctname/repo like `ocelot summary` command~~
~~- change client's --validate command to take in a value just like all the other commands~~
~~- hash matching only works when you pass in full hash~~
    ~~- when printing matching git hashes, should display corresponding acctname/repo like `ocelot summary` command~~
