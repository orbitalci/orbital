package main

import (
    "bytes"
    "github.com/ghodss/yaml"
    "github.com/golang/protobuf/proto"
    lg "github.com/shankj3/ocelot/ocelog"
    "io/ioutil"
)

// really hack-y way to convert yml to protobuf message.
// converts yml to json byte array, then uses HandleUnmarshal to
// convert json to the proto.Message `msg` returns error.
func ConvertYAMLtoProtobuf(yml []byte, msg proto.Message) error {
    json_bytes, err := yaml.YAMLToJSON(yml)
    if err != nil {
        lg.Log().Warn("couldn't marshal yml to json")
        return err
    }
    HandleUnmarshal(ioutil.NopCloser(bytes.NewReader(json_bytes)), msg)
    lg.Log().Debug(msg)
    return nil
}
