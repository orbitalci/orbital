package main

import (
    // "github.com/golang/protobuf/proto"
    "github.com/shankj3/ocelot/ocelog"
    pb "github.com/shankj3/ocelot/protos"
    "gopkg.in/go-playground/webhooks.v3/bitbucket"
    // "time"
)

func ConvertHookToProto(repopush bitbucket.RepoPushPayload) pb.RepoPush {
    // todo: this is dumb, just unmarshall directly to protobuf
    ocelog.Log.Debug("Converting Repo Push to pb.Repopush proto struct")
    latestChange := repopush.Push.Changes[0]
    pushHook := pb.RepoPush{
        User:             repopush.Actor.Username,
        RepoLink:         repopush.Repository.Links.HTML.Href,
        LastHash:         latestChange.New.Target.Hash,
        IdentifierType:   latestChange.New.Type,
        IdentifierString: latestChange.New.Name,
        CommitTime:       latestChange.New.Target.Date.Unix(),
    }
    return pushHook
}
