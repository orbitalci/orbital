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