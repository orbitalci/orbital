# project ocelot


make something that isn't fancy but works good  

## What do we want out of it?  

*ability to add extra workers and use them to distribute load*
Use a Queueing system (nsq? go doesn't really need a special queing system, can just set one up via channels)

*yml configuration in repository*   
*webhook triggers*  
*no special configuration on machine required, no "snowflakes"*  
*build in docker containers*  
*have endpoints on build number to allow for checking status of build*
*simple notification system*
*simple website that shows status of builds*
*endpoint for getting historical data*
*prometheus exporter for those stats (or something like that)*

## How do I get started?

Since the golang code is built inside the container, you only need to have a docker host to run containers, and docker-compose installed (Compose file format 3.3).

* docker-engine >= 17.06.0+
* docker-compose >= 1.14.0+

### Build

To build, you should run `./build-release.sh` from the project root.

The build occurs in 3 steps.
1. We build a base build image `ocelot-build` that includes all the tools, and golang library dependencies we need to compiling the project. 
2. Using `ocelot-build`, we statically compile the golang binaries.
3. Using docker [multi-stage build](https://docs.docker.com/engine/userguide/eng-image/multistage-build/#use-multi-stage-builds) features, we copy the compiled-binary into an empty [scratch](https://hub.docker.com/_/scratch/) container.

#### Troubleshooting build
If the build is acting funny, you can use this `docker run` command to start a shell in the builder environment.

    docker run --rm -it \
    -v $(pwd):/go/src/github.com/shankj3/ocelot/ \
    -e SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" \
    ocelot-build sh

### Run

To start a local development cluster:

From `${OCELOT_ROOT}` or `${OCELOT_ROOT}/${SERVICE_ROOT}` run the following to run every service (including infrastructure):

`docker-compose up` to run everything in the foreground

`docker-compose up -d` to run everything in the background

`docker-compose stop` to stop it all

## Known issues:
* Starting an Ocelot development cluster with `docker-compose up` does not result in a fully wired, functional system.

## Proposed features
 * store length of builds
 * monitoring process??   
 * resource management  
 * support different failure conditions
   * re-queueing vs error reporting 
 * define trigger section for build?

## Target projects
 * Ocelot
 * orchestr8 - The reference java project

## TODO:
 * Fix firewall rules (w/ Tanner's assistance) so we can route webhooks internally
 * Deploy onto VM

marianne: grpc admin & hookhandler remote config
jessi: grpc streaming & changes from PR & converting ocelot.yml -> abbys pipeline proto message
abby: pipeline stuff | docker / kubernetes implementation | client related stuff 
tj: building & running project & registrator alternative? or registrator? 


# what to post to admin for creds 

```{
   "configId":"jessishank", 
   "clientId": "QEBYwP5cKAC3ykhau4",
   "clientSecret": "secretz",
   "tokenURL": "https://bitbucket.org/site/oauth2/access_token", 
   "acctName": "jessishank",
   "type": "bitbucket"
}```



## flow for subscriptions 
1. user commits to master of tracked branch (repoName: task-service) the subscriptions block of their ocelot.yml: 
    subscriptions:
    - acctRepo: level11consulting/pogo
      alias: pogo
      branchQueueMap: 
      - master:master
      - develop:develop
2. when master is built, this subscription is picked up and stored in active_subscriptions table 
3. task service continues to be committed to, and when it does normal build activities occur with no subscrition-related activity 
4. level11consulting/pogo builds successfully in ocelot. 
5. in the postFlight section of launcher for the most recent level11consulting/pogo build, where pr comments / notifications could occur, there is a check in the active subscriptions table.
   - if there is no entry in the active_subscriptions table where subscribed_to_acct_repo='level11consulting/pogo', then nothing happens. 
   - If there IS an entry:
        - a new build will be queued for level11consulting/task-service, with its signaledBy being SUBSCRIBED
        - new row in subscription_data created that associates build id of level11consulting/task-service with the build id of level11consulting/pogo

builds triggered by subscriptions will be be treated largely as normal, except they will have more environment variables
