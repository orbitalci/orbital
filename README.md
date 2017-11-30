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

From `${OCELOT_ROOT}` or `${OCELOT_ROOT}/${SERVICE_ROOT}` run the following to build every service:
`docker-compose build`

(There should be a `${OCELOT_ROOT}/${SERVICE_ROOT}/docker-compose.yml` that sets the build context to `${OCELOT_ROOT}`, because the build context includes vendoring dependencies from `dep`)

### Run

To start a local development cluster:

From `${OCELOT_ROOT}` or `${OCELOT_ROOT}/${SERVICE_ROOT}` run the following to run every service (including infrastructure):

`docker-compose up` to run everything in the foreground

`docker-compose up -d` to run everything in the background

`docker-compose stop` to stop it all

## Known issues:
* Starting an Ocelot development cluster with `docker-compose up` does not result in a fully wired, functional system.

store length of builds


monitoring process??   
resource management  
support different failure conditions?   
re-queueing vs error reporting 

*define trigger section for build?* 

# TODO: TANNER HOLE PUNCH

build ocelot is for us

## java test project  
orchestr8



#DEPLOY ON VM! 


marianne: grpc admin & hookhandler remote config
jessi: grpc streaming & changes from PR & converting ocelot.yml -> abbys pipeline proto message
abby: pipeline stuff | docker / kubernetes implementation | client related stuff 
tj: building & running project & registrator alternative? or registrator? 
>>>>>>> master
