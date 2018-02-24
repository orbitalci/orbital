# more code more problems

## Questions
- one repo, multiple pipelines? how to handle that 
    - for example, the docker builder pattern. when deps are updated, the docker base image shoudl be as well. that's all in one repo, but its widely different behavior than build/testing source code.

## in progress
- remove volume mounts on spawned build containers - they should be downloading the bash scripts out of s3 
- make it so that we can query whether or not there's a key uploaded for this accountname
- `ocelot status` - asks admin to get build runtime (marianne) 
    ~~- current / past stage info to be added to build_stage_details (currently build_failure_reason)~~
    - this should really give you success/failure of some sort
        - if running or failed, display stages + corresponding messages 
    - queryable by:
        - repoName 
        - acctname/repoName
        ~~- git hash~~
- sweep through repo and add updates to db at any point of failure; some areas that i can think of off hand: (jessi)                                     
    - werker dies... should update that hash somehow with build failure reason -> dead werker (at least, possibly also a re-queue)
         - panic recovery on main function 
            - cleanup consul entry / notify _someone_ of status 
            - [1](https://blog.golang.org/defer-panic-and-recover), [2](https://golangbot.com/panic-and-recover/)
            - cleanup docker containers
            - add item back to queue for build 
    
## bugs: 
~~- GOTTA FIGURE OUT WHAT TO DO ABOUT HANDLING SSH KEYS~~
    ~~- take in a private key via command line for an account name~~
- be able to properly handle KILLS (what happens when build is killed halfway?)
    - [stackoverflow](https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in-a-defe)
- fix dev mode (maybe fuck this)
- fix html viewer displaying *oldest* matching git hash (maybe fuck this)     

## BIG TODOs:
- actually parse out exit codes, not just shit itself if it gets a non-zero one
- put a limit on number of running containers at once
- failure notifications
- actions to only take based on branch or w/e 
    - solution: add `trigger/skip` section to stage yml?
        - implementation: when hookhandler receives a message, it will filter on new commit messages and branch to take out stages that do not fit trigger requirements
- tag built projects by custom group name, that way you can filter to see all repos belonging to your group
- docker login? - our repo creds model works for this currently, just need to implement part that actually runs docker login
- do the pipeline thing
- make it so that successful builds will edit the commit message and you can see in commit history whether or not that build was successful 
- tighter maven integration?
- `ocelot kill <hash>` - add a quit channel
- check out worker queue
	- storing how long your commit waited on the queue
- return to old package install list??
    - [docker get script for ex](https://get.docker.com/)    

    
## little TODOs?:
- something that says X isn't tracked by ocelot (ADD THIS CHECK TO ALL COMMANDS SO THAT BEHAVIOR IS CONSISTENT) 
- something to take care of removing dead docker containers + images from werker's host (this shit builds up fast)
- add ability to remove webhooks??? (this would be handy while we're playing around with stuff)
- add ability to specify if you want all branches built
- add optional to specify working directory inside of ocelot.yml
- `ocelot watch` - create a new webhook
- `ocelot trigger jessishank/mytestocy <hash>` - to put on queue w/o bitbucket webhook
- polling option to add new repos with ocelot.yml
- `ocelot summary` command takes in -repo or --hash? 
- add `who triggered this` value to summary table
- add *what* triggered this build
	- command line trigger
	- pr trigger
	- commit trigger
	- etc. 

## done:
- ~~fix `ocelot logs ` retrieval by build_id~~
- ~~nexus (jessi)~~
- ~~assume that the git = whatever you last pushed up (run `git rev-parse` command) ~~
- ~~fix goddamn tests~~ 
- ~~store stages to db~~
- ~~when printing matching git hashes, should display corresponding acctname/repo like `ocelot summary` command~~
- ~~change client's --validate command to take in a value just like all the other commands~~
- ~~hash matching only works when you pass in full hash~~
    - ~~when printing matching git hashes, should display corresponding acctname/repo like `ocelot summary` command~~
- ~~failed validation at hook handler stage~~ 
- ~~failed setup stage~~
    - ~~should probably start printing out the actual docker errors...~~
- ~~more verbose for other stage failures --  already done by marianne?~~            
