// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: build.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	build.proto
	guideocelot.proto
	werkerserver.proto

It has these top-level messages:
	BuildConfig
	Stage
	Result
	Triggers
	WerkerTask
	BuildReq
	AllCredsWrapper
	CredWrapper
	SSHKeyWrapper
	SSHWrap
	VCSCreds
	RepoCredWrapper
	RepoCreds
	K8SCreds
	K8SCredsWrapper
	StatusQuery
	BuildQuery
	Builds
	BuildRuntimeInfo
	LineResponse
	RepoAccount
	Status
	StageStatus
	BuildSummary
	Summaries
	PollRequest
	Polls
	Exists
	Request
	Response
*/
package pb

import go_proto_validators "github.com/mwitkow/go-proto-validators"
import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *BuildConfig) Validate() error {
	for _, item := range this.Stages {
		if item != nil {
			if err := go_proto_validators.CallValidatorIfExists(item); err != nil {
				return go_proto_validators.FieldError("Stages", err)
			}
		}
	}
	return nil
}
func (this *Stage) Validate() error {
	if this.Trigger != nil {
		if err := go_proto_validators.CallValidatorIfExists(this.Trigger); err != nil {
			return go_proto_validators.FieldError("Trigger", err)
		}
	}
	return nil
}
func (this *Result) Validate() error {
	return nil
}
func (this *Triggers) Validate() error {
	return nil
}
func (this *WerkerTask) Validate() error {
	if this.BuildConf != nil {
		if err := go_proto_validators.CallValidatorIfExists(this.BuildConf); err != nil {
			return go_proto_validators.FieldError("BuildConf", err)
		}
	}
	return nil
}
