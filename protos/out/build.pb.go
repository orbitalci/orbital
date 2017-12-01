// Code generated by protoc-gen-go. DO NOT EDIT.
// source: build.proto

/*
Package protos is a generated protocol buffer package.

It is generated from these files:
	build.proto
	common.proto
	commonevententities.proto
	projectrootdir.proto
	respositories.proto
	webhook.proto

It has these top-level messages:
	BuildConfig
	Stage
	PushBuildBundle
	PRBuildBundle
	LinkUrl
	LinkAndName
	Links
	Owner
	Repository
	PullRequestEntity
	PRInfo
	Project
	Changeset
	Commit
	RepoSourceFile
	PaginatedRootDirs
	PaginatedRepository
	RepoPush
	PullRequest
	PullRequestApproved
	CreateWebhook
	GetWebhooks
	Webhooks
*/
package protos

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BuildConfig struct {
	Image          string   `protobuf:"bytes,1,opt,name=image" json:"image,omitempty"`
	DockerPackages []string `protobuf:"bytes,2,rep,name=dockerPackages" json:"dockerPackages,omitempty"`
	BeforeStages   *Stage   `protobuf:"bytes,3,opt,name=beforeStages" json:"beforeStages,omitempty"`
	AfterStages    *Stage   `protobuf:"bytes,4,opt,name=afterStages" json:"afterStages,omitempty"`
	Build          *Stage   `protobuf:"bytes,5,opt,name=build" json:"build,omitempty"`
	Test           *Stage   `protobuf:"bytes,6,opt,name=test" json:"test,omitempty"`
	Deploy         *Stage   `protobuf:"bytes,7,opt,name=deploy" json:"deploy,omitempty"`
}

func (m *BuildConfig) Reset()                    { *m = BuildConfig{} }
func (m *BuildConfig) String() string            { return proto.CompactTextString(m) }
func (*BuildConfig) ProtoMessage()               {}
func (*BuildConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BuildConfig) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *BuildConfig) GetDockerPackages() []string {
	if m != nil {
		return m.DockerPackages
	}
	return nil
}

func (m *BuildConfig) GetBeforeStages() *Stage {
	if m != nil {
		return m.BeforeStages
	}
	return nil
}

func (m *BuildConfig) GetAfterStages() *Stage {
	if m != nil {
		return m.AfterStages
	}
	return nil
}

func (m *BuildConfig) GetBuild() *Stage {
	if m != nil {
		return m.Build
	}
	return nil
}

func (m *BuildConfig) GetTest() *Stage {
	if m != nil {
		return m.Test
	}
	return nil
}

func (m *BuildConfig) GetDeploy() *Stage {
	if m != nil {
		return m.Deploy
	}
	return nil
}

type Stage struct {
	Env    []string `protobuf:"bytes,1,rep,name=env" json:"env,omitempty"`
	Script []string `protobuf:"bytes,2,rep,name=script" json:"script,omitempty"`
}

func (m *Stage) Reset()                    { *m = Stage{} }
func (m *Stage) String() string            { return proto.CompactTextString(m) }
func (*Stage) ProtoMessage()               {}
func (*Stage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Stage) GetEnv() []string {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *Stage) GetScript() []string {
	if m != nil {
		return m.Script
	}
	return nil
}

type PushBuildBundle struct {
	Config       *BuildConfig `protobuf:"bytes,1,opt,name=config" json:"config,omitempty"`
	PushData     *RepoPush    `protobuf:"bytes,2,opt,name=pushData" json:"pushData,omitempty"`
	VaultToken   string       `protobuf:"bytes,3,opt,name=vaultToken" json:"vaultToken,omitempty"`
	CheckoutHash string       `protobuf:"bytes,4,opt,name=checkoutHash" json:"checkoutHash,omitempty"`
}

func (m *PushBuildBundle) Reset()                    { *m = PushBuildBundle{} }
func (m *PushBuildBundle) String() string            { return proto.CompactTextString(m) }
func (*PushBuildBundle) ProtoMessage()               {}
func (*PushBuildBundle) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PushBuildBundle) GetConfig() *BuildConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *PushBuildBundle) GetPushData() *RepoPush {
	if m != nil {
		return m.PushData
	}
	return nil
}

func (m *PushBuildBundle) GetVaultToken() string {
	if m != nil {
		return m.VaultToken
	}
	return ""
}

func (m *PushBuildBundle) GetCheckoutHash() string {
	if m != nil {
		return m.CheckoutHash
	}
	return ""
}

type PRBuildBundle struct {
	Config       *BuildConfig `protobuf:"bytes,1,opt,name=config" json:"config,omitempty"`
	PrData       *PullRequest `protobuf:"bytes,2,opt,name=prData" json:"prData,omitempty"`
	VaultToken   string       `protobuf:"bytes,3,opt,name=vaultToken" json:"vaultToken,omitempty"`
	CheckoutHash string       `protobuf:"bytes,4,opt,name=checkoutHash" json:"checkoutHash,omitempty"`
}

func (m *PRBuildBundle) Reset()                    { *m = PRBuildBundle{} }
func (m *PRBuildBundle) String() string            { return proto.CompactTextString(m) }
func (*PRBuildBundle) ProtoMessage()               {}
func (*PRBuildBundle) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PRBuildBundle) GetConfig() *BuildConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *PRBuildBundle) GetPrData() *PullRequest {
	if m != nil {
		return m.PrData
	}
	return nil
}

func (m *PRBuildBundle) GetVaultToken() string {
	if m != nil {
		return m.VaultToken
	}
	return ""
}

func (m *PRBuildBundle) GetCheckoutHash() string {
	if m != nil {
		return m.CheckoutHash
	}
	return ""
}

func init() {
	proto.RegisterType((*BuildConfig)(nil), "protos.BuildConfig")
	proto.RegisterType((*Stage)(nil), "protos.Stage")
	proto.RegisterType((*PushBuildBundle)(nil), "protos.PushBuildBundle")
	proto.RegisterType((*PRBuildBundle)(nil), "protos.PRBuildBundle")
}

func init() { proto.RegisterFile("build.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 356 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0x51, 0x4e, 0x83, 0x40,
	0x10, 0x86, 0x43, 0x5b, 0x50, 0x86, 0x56, 0x9b, 0xd5, 0x18, 0xe2, 0x83, 0xa9, 0x18, 0x4d, 0x93,
	0x9a, 0x9a, 0xd6, 0x1b, 0x54, 0x1f, 0x7c, 0x24, 0xab, 0x17, 0x58, 0x60, 0x5a, 0x08, 0xc8, 0x22,
	0xbb, 0x5b, 0xe3, 0x65, 0xbc, 0x81, 0x89, 0x47, 0x34, 0x5d, 0x68, 0x43, 0x0d, 0x4f, 0xc6, 0xa7,
	0xdd, 0xf9, 0xe7, 0x9b, 0x9d, 0x99, 0x3f, 0x0b, 0x4e, 0xa0, 0x92, 0x2c, 0x9a, 0x16, 0x25, 0x97,
	0x9c, 0x58, 0xfa, 0x10, 0xe7, 0x83, 0x77, 0x0c, 0x62, 0xce, 0xd3, 0x4a, 0xf6, 0x3e, 0x3b, 0xe0,
	0x2c, 0x36, 0xd8, 0x03, 0xcf, 0x97, 0xc9, 0x8a, 0x9c, 0x82, 0x99, 0xbc, 0xb2, 0x15, 0xba, 0xc6,
	0xc8, 0x18, 0xdb, 0xb4, 0x0a, 0xc8, 0x0d, 0x1c, 0x45, 0x3c, 0x4c, 0xb1, 0xf4, 0x59, 0x98, 0xb2,
	0x15, 0x0a, 0xb7, 0x33, 0xea, 0x8e, 0x6d, 0xfa, 0x4b, 0x25, 0x33, 0xe8, 0x07, 0xb8, 0xe4, 0x25,
	0x3e, 0x4b, 0x4d, 0x75, 0x47, 0xc6, 0xd8, 0x99, 0x0f, 0xaa, 0x5e, 0x62, 0xaa, 0x55, 0xba, 0x87,
	0x90, 0x3b, 0x70, 0xd8, 0x52, 0x62, 0x59, 0x57, 0xf4, 0xda, 0x2a, 0x9a, 0x04, 0xb9, 0x02, 0x53,
	0xef, 0xe5, 0x9a, 0x6d, 0x68, 0x95, 0x23, 0x97, 0xd0, 0x93, 0x28, 0xa4, 0x6b, 0xb5, 0x31, 0x3a,
	0x45, 0xae, 0xc1, 0x8a, 0xb0, 0xc8, 0xf8, 0x87, 0x7b, 0xd0, 0x06, 0xd5, 0x49, 0x6f, 0x06, 0xa6,
	0x16, 0xc8, 0x10, 0xba, 0x98, 0xaf, 0x5d, 0x43, 0x2f, 0xbe, 0xb9, 0x92, 0x33, 0xb0, 0x44, 0x58,
	0x26, 0x85, 0xac, 0xdd, 0xa8, 0x23, 0xef, 0xdb, 0x80, 0x63, 0x5f, 0x89, 0x58, 0xfb, 0xba, 0x50,
	0x79, 0x94, 0x21, 0x99, 0x80, 0x15, 0x6a, 0x87, 0xb5, 0xb1, 0xce, 0xfc, 0x64, 0xdb, 0xad, 0x61,
	0x3e, 0xad, 0x11, 0x72, 0x0b, 0x87, 0x85, 0x12, 0xf1, 0x23, 0x93, 0xcc, 0xed, 0x68, 0x7c, 0xb8,
	0xc5, 0x29, 0x16, 0x7c, 0xf3, 0x36, 0xdd, 0x11, 0xe4, 0x02, 0x60, 0xcd, 0x54, 0x26, 0x5f, 0x78,
	0x8a, 0xb9, 0xb6, 0xdc, 0xa6, 0x0d, 0x85, 0x78, 0xd0, 0x0f, 0x63, 0x0c, 0x53, 0xae, 0xe4, 0x13,
	0x13, 0xb1, 0xb6, 0xd8, 0xa6, 0x7b, 0x9a, 0xf7, 0x65, 0xc0, 0xc0, 0xa7, 0x7f, 0x1e, 0x78, 0x02,
	0x56, 0x51, 0x36, 0xc6, 0xdd, 0xc1, 0xbe, 0xca, 0x32, 0x8a, 0x6f, 0x0a, 0x85, 0xa4, 0x35, 0xf2,
	0x1f, 0xf3, 0x06, 0xd5, 0x6f, 0xbe, 0xff, 0x09, 0x00, 0x00, 0xff, 0xff, 0xfe, 0x48, 0xd8, 0x53,
	0xe3, 0x02, 0x00, 0x00,
}
