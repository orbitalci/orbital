#!/usr/bin/env python3

import sys
import json
import requests


CRED_URL = 'http://{host_port}/v1/creds/{credtype}'
GET_UPDATE_URL = CRED_URL + '/{account}/{identifier}?subType={st}'

def migrate(dumped_creds_loc, host_port):
    """
    migrate will take a credential file like the one shown at the bottom of this script 
    and upload it to a new ocelot instance to migrate all credentials quickly.
    """
    with open(dumped_creds_loc) as all_creddies:
        all_loaded = json.load(all_creddies)
    for cred_t, creds in all_loaded.items():
        print("\n====================================\n")
        print("PROCESSING " + cred_t)
        print("\n====================================\n")
        for cred in creds: 
            # overwrite notify nastiboi
            cred_t = 'notify' if cred_t == "NOTIFIER" else cred_t
            print("handling %s|%s" %(cred["acctName"], cred["identifier"]))
            # check if cred already exists
            get_update_url = GET_UPDATE_URL.format(host_port=host_port, credtype=cred_t.lower(), account=cred["acctName"], st=cred["subType"], identifier=cred["identifier"])
            get = requests.get(get_update_url)
            # if it does, update
            if get.ok:
                print("updating existing credential")
                upd = requests.put(get_update_url, json=cred)
                if not upd.ok:
                    print("failed to update credential, error is %s" % upd.text)
                else:
                    print("succesfully updated")
            # if not, make a new cred
            else:
                print("creating new credential")
                create_url = CRED_URL.format(host_port=host_port, credtype=cred_t.lower())
                cre = requests.post(create_url, json=cred)
                if not cre.ok:
                    print("failed to create credential, error is %s" % cre.text)
                else:
                    print("succesfully created")
            print("\n--------\n")


if __name__ == "__main__":
    if sys.argv[1] in ["help", "--help", "-help"]:
        print("migrate a dumped ocelot creds file to a new ocelot instance. credfile loc must be first argument, host/port or name is second")
        sys.exit(0)

    try:
        dump_loc, host_port = sys.argv[1], sys.argv[2]
    except:
        print("must provide dumped creds location and host port as positional arguments. credfile is first, host_port is second!")
        sys.exit(1)
    
    migrate(dump_loc, host_port)



"""
EXAMPLE CREDENTIAL FILE
{
  "K8S": [
    {
      "acctName": "shankj3",
      "k8sContents": "<<REDACTED>>",
      "identifier": "config",
      "subType": "KUBECONF"
    }
  ],
  "NOTIFIER": [
    {
      "acctName": "shankj3",
      "subType": "SLACK",
      "identifier": "L11_SLACK",
      "clientSecret": "<<REDACTED>>"
    },
    {
      "acctName": "jessishank",
      "subType": "SLACK",
      "identifier": "private_slack",
      "clientSecret": "<<REDACTED>>"
    }
  ],
  "REPO": [
    {
      "username": "admin",
      "password": "<<REDACTED>>",
      "repoUrl": "https://insecure-nexus:9028/repository/maven-snapshots/",
      "identifier": "maven-snapshots",
      "acctName": "shankj3",
      "subType": "NEXUS"
    },
    {
      "username": "admin",
      "password": "<<REDACTED>>",
      "repoUrl": "insecure-docker.com:443",
      "identifier": "docker-nexus",
      "acctName": "level11consulting",
      "subType": "DOCKER"
    }
  ],
  "SSH": [
    {
      "acctName": "shankj3",
      "privateKey": "<<REDACTED>>",
      "subType": "SSHKEY",
      "identifier": "RSA_KEY"
    }
  ],
  "VCS": [
    {
      "clientId": "VEhMhdw6uprevzh8Du",
      "clientSecret": "<<REDACTED>>",
      "identifier": "BITBUCKET_shankj3",
      "tokenURL": "https://bitbucket.org/site/oauth2/access_token",
      "acctName": "shankj3",
      "subType": "BITBUCKET"
    }
  ]
}
"""