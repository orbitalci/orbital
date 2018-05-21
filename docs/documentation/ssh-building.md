## lessons learned

the machine that will be ssh-ed to for build execution needs these things modified: 
 - `/etc/ssh/sshd_config`
    - `AcceptEnv LANG *`
        - this is required so that you can pass environment variables to the ssh session
    - `PermitUserEnvironment yes`
        - see above reason 
- `echo "PATH=/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/sbin:/usr/local/bin" >> ~/.ssh/environment` 
    - at least for mac, not all the normal bin paths are put on the ssh session's path, so you have to add a few and put it in the `~/.ssh` directory of the user you will be connecting as
    