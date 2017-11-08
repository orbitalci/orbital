/*
Worker needs to:

Pull off of NSQ Queue
Process config file
run build in docker container
provide results endpoint, way for server to access data
*/

package main

import (
    "github.com/golang/protobuf/proto"
    // "github.com/shankj3/ocelot/dockering"
    "github.com/shankj3/ocelot/nsqpb"
    "github.com/shankj3/ocelot/ocelog"
    pb "github.com/shankj3/ocelot/protos/out"
)

func HandleRepoPushMessage(message []byte) error {
    push := &pb.RepoPush{}
    ocelog.Log().Debug("hit HandleRepoPush")
    if err := proto.Unmarshal(message, push); err != nil {
        ocelog.LogErrField(err).Warning("unmarshal error")
        return err
    }
    go Build(push)
    return nil
}

func Build(buildjob *pb.RepoPush) error {
    ocelog.Log().Info(buildjob.GetActor().GetUsername())
    return nil
}

func main() {
    ocelog.InitializeOcelog(ocelog.GetFlags())
    protoConsume := &nsqpb.ProtoConsume{}
    protoConsume.UnmarshalProtoFunc = HandleRepoPushMessage
    nsqpb.ConsumeMessages(protoConsume, "repo_push", "one")
}
