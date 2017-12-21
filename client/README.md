CLI actions
===


`ocelot status`
---
If sent with any flags, it will retrieve the status of all running builds.   
*Additional Flags*:  
- `ocelot status -account=<account_name>`: all jobs associated with `<account_name>`  
- `ocelot status -job=<hash>`: status of singular job w/ id `<hash>`  


`ocelot logs`
---
I only really see it in one usecase, which would be tailing the logs of a specific build.
Probably will end up looking like `ocelot logs <hash>`


`ocelot creds`
---
- `ocelot creds list`: list accounts 
- `ocelot creds add`: add account info (flags or stdin)


`ocelot integrations`
---
Adding in creds for integrations 

haven't fully fleshed out this idea, but thinking `ocelot integrations add <file_name>?`