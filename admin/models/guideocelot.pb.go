// Code generated by protoc-gen-go. DO NOT EDIT.
// source: guideocelot.proto

/*
Package models is a generated protocol buffer package.

It is generated from these files:
	guideocelot.proto

It has these top-level messages:
	AllCredsWrapper
	CredWrapper
	VCSCreds
	RepoCredWrapper
	RepoCreds
	BuildQuery
	BuildRuntimeInfo
	LogResponse
	RepoAccount
	BuildSummary
	Summaries
*/
package models

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/empty"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import google_protobuf2 "github.com/golang/protobuf/ptypes/timestamp"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AllCredsWrapper struct {
	RepoCreds *RepoCredWrapper `protobuf:"bytes,1,opt,name=repoCreds" json:"repoCreds,omitempty"`
	VcsCreds  *CredWrapper     `protobuf:"bytes,3,opt,name=vcsCreds" json:"vcsCreds,omitempty"`
}

func (m *AllCredsWrapper) Reset()                    { *m = AllCredsWrapper{} }
func (m *AllCredsWrapper) String() string            { return proto.CompactTextString(m) }
func (*AllCredsWrapper) ProtoMessage()               {}
func (*AllCredsWrapper) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AllCredsWrapper) GetRepoCreds() *RepoCredWrapper {
	if m != nil {
		return m.RepoCreds
	}
	return nil
}

func (m *AllCredsWrapper) GetVcsCreds() *CredWrapper {
	if m != nil {
		return m.VcsCreds
	}
	return nil
}

type CredWrapper struct {
	Vcs []*VCSCreds `protobuf:"bytes,2,rep,name=vcs" json:"vcs,omitempty"`
}

func (m *CredWrapper) Reset()                    { *m = CredWrapper{} }
func (m *CredWrapper) String() string            { return proto.CompactTextString(m) }
func (*CredWrapper) ProtoMessage()               {}
func (*CredWrapper) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CredWrapper) GetVcs() []*VCSCreds {
	if m != nil {
		return m.Vcs
	}
	return nil
}

type VCSCreds struct {
	ClientId     string `protobuf:"bytes,1,opt,name=clientId" json:"clientId,omitempty"`
	ClientSecret string `protobuf:"bytes,2,opt,name=clientSecret" json:"clientSecret,omitempty"`
	TokenURL     string `protobuf:"bytes,3,opt,name=tokenURL" json:"tokenURL,omitempty"`
	AcctName     string `protobuf:"bytes,4,opt,name=acctName" json:"acctName,omitempty"`
	Type         string `protobuf:"bytes,5,opt,name=type" json:"type,omitempty"`
}

func (m *VCSCreds) Reset()                    { *m = VCSCreds{} }
func (m *VCSCreds) String() string            { return proto.CompactTextString(m) }
func (*VCSCreds) ProtoMessage()               {}
func (*VCSCreds) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *VCSCreds) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *VCSCreds) GetClientSecret() string {
	if m != nil {
		return m.ClientSecret
	}
	return ""
}

func (m *VCSCreds) GetTokenURL() string {
	if m != nil {
		return m.TokenURL
	}
	return ""
}

func (m *VCSCreds) GetAcctName() string {
	if m != nil {
		return m.AcctName
	}
	return ""
}

func (m *VCSCreds) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

type RepoCredWrapper struct {
	Repo []*RepoCreds `protobuf:"bytes,3,rep,name=repo" json:"repo,omitempty"`
}

func (m *RepoCredWrapper) Reset()                    { *m = RepoCredWrapper{} }
func (m *RepoCredWrapper) String() string            { return proto.CompactTextString(m) }
func (*RepoCredWrapper) ProtoMessage()               {}
func (*RepoCredWrapper) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *RepoCredWrapper) GetRepo() []*RepoCreds {
	if m != nil {
		return m.Repo
	}
	return nil
}

type RepoCreds struct {
	Username string `protobuf:"bytes,1,opt,name=username" json:"username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
	RepoUrl  string `protobuf:"bytes,3,opt,name=repoUrl" json:"repoUrl,omitempty"`
	AcctName string `protobuf:"bytes,4,opt,name=acctName" json:"acctName,omitempty"`
	Type     string `protobuf:"bytes,5,opt,name=type" json:"type,omitempty"`
}

func (m *RepoCreds) Reset()                    { *m = RepoCreds{} }
func (m *RepoCreds) String() string            { return proto.CompactTextString(m) }
func (*RepoCreds) ProtoMessage()               {}
func (*RepoCreds) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *RepoCreds) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *RepoCreds) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *RepoCreds) GetRepoUrl() string {
	if m != nil {
		return m.RepoUrl
	}
	return ""
}

func (m *RepoCreds) GetAcctName() string {
	if m != nil {
		return m.AcctName
	}
	return ""
}

func (m *RepoCreds) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

type BuildQuery struct {
	Hash    string `protobuf:"bytes,1,opt,name=hash" json:"hash,omitempty"`
	BuildId int64  `protobuf:"varint,2,opt,name=buildId" json:"buildId,omitempty"`
}

func (m *BuildQuery) Reset()                    { *m = BuildQuery{} }
func (m *BuildQuery) String() string            { return proto.CompactTextString(m) }
func (*BuildQuery) ProtoMessage()               {}
func (*BuildQuery) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *BuildQuery) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *BuildQuery) GetBuildId() int64 {
	if m != nil {
		return m.BuildId
	}
	return 0
}

type BuildRuntimeInfo struct {
	Done     bool   `protobuf:"varint,1,opt,name=done" json:"done,omitempty"`
	Ip       string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
	GrpcPort string `protobuf:"bytes,3,opt,name=grpcPort" json:"grpcPort,omitempty"`
}

func (m *BuildRuntimeInfo) Reset()                    { *m = BuildRuntimeInfo{} }
func (m *BuildRuntimeInfo) String() string            { return proto.CompactTextString(m) }
func (*BuildRuntimeInfo) ProtoMessage()               {}
func (*BuildRuntimeInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *BuildRuntimeInfo) GetDone() bool {
	if m != nil {
		return m.Done
	}
	return false
}

func (m *BuildRuntimeInfo) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *BuildRuntimeInfo) GetGrpcPort() string {
	if m != nil {
		return m.GrpcPort
	}
	return ""
}

type LogResponse struct {
	OutputLine string `protobuf:"bytes,1,opt,name=outputLine" json:"outputLine,omitempty"`
}

func (m *LogResponse) Reset()                    { *m = LogResponse{} }
func (m *LogResponse) String() string            { return proto.CompactTextString(m) }
func (*LogResponse) ProtoMessage()               {}
func (*LogResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *LogResponse) GetOutputLine() string {
	if m != nil {
		return m.OutputLine
	}
	return ""
}

type RepoAccount struct {
	Repo    string `protobuf:"bytes,1,opt,name=repo" json:"repo,omitempty"`
	Account string `protobuf:"bytes,2,opt,name=account" json:"account,omitempty"`
	Limit   int32  `protobuf:"varint,3,opt,name=limit" json:"limit,omitempty"`
}

func (m *RepoAccount) Reset()                    { *m = RepoAccount{} }
func (m *RepoAccount) String() string            { return proto.CompactTextString(m) }
func (*RepoAccount) ProtoMessage()               {}
func (*RepoAccount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *RepoAccount) GetRepo() string {
	if m != nil {
		return m.Repo
	}
	return ""
}

func (m *RepoAccount) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *RepoAccount) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type BuildSummary struct {
	Hash          string                      `protobuf:"bytes,1,opt,name=hash" json:"hash,omitempty"`
	Failed        bool                        `protobuf:"varint,2,opt,name=failed" json:"failed,omitempty"`
	BuildTime     *google_protobuf2.Timestamp `protobuf:"bytes,3,opt,name=buildTime" json:"buildTime,omitempty"`
	Account       string                      `protobuf:"bytes,4,opt,name=account" json:"account,omitempty"`
	BuildDuration float64                     `protobuf:"fixed64,5,opt,name=buildDuration" json:"buildDuration,omitempty"`
	Repo          string                      `protobuf:"bytes,6,opt,name=repo" json:"repo,omitempty"`
	Branch        string                      `protobuf:"bytes,7,opt,name=branch" json:"branch,omitempty"`
	BuildId       int64                       `protobuf:"varint,8,opt,name=buildId" json:"buildId,omitempty"`
}

func (m *BuildSummary) Reset()                    { *m = BuildSummary{} }
func (m *BuildSummary) String() string            { return proto.CompactTextString(m) }
func (*BuildSummary) ProtoMessage()               {}
func (*BuildSummary) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *BuildSummary) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *BuildSummary) GetFailed() bool {
	if m != nil {
		return m.Failed
	}
	return false
}

func (m *BuildSummary) GetBuildTime() *google_protobuf2.Timestamp {
	if m != nil {
		return m.BuildTime
	}
	return nil
}

func (m *BuildSummary) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *BuildSummary) GetBuildDuration() float64 {
	if m != nil {
		return m.BuildDuration
	}
	return 0
}

func (m *BuildSummary) GetRepo() string {
	if m != nil {
		return m.Repo
	}
	return ""
}

func (m *BuildSummary) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

func (m *BuildSummary) GetBuildId() int64 {
	if m != nil {
		return m.BuildId
	}
	return 0
}

type Summaries struct {
	Sums []*BuildSummary `protobuf:"bytes,1,rep,name=sums" json:"sums,omitempty"`
}

func (m *Summaries) Reset()                    { *m = Summaries{} }
func (m *Summaries) String() string            { return proto.CompactTextString(m) }
func (*Summaries) ProtoMessage()               {}
func (*Summaries) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *Summaries) GetSums() []*BuildSummary {
	if m != nil {
		return m.Sums
	}
	return nil
}

func init() {
	proto.RegisterType((*AllCredsWrapper)(nil), "models.AllCredsWrapper")
	proto.RegisterType((*CredWrapper)(nil), "models.CredWrapper")
	proto.RegisterType((*VCSCreds)(nil), "models.VCSCreds")
	proto.RegisterType((*RepoCredWrapper)(nil), "models.RepoCredWrapper")
	proto.RegisterType((*RepoCreds)(nil), "models.RepoCreds")
	proto.RegisterType((*BuildQuery)(nil), "models.BuildQuery")
	proto.RegisterType((*BuildRuntimeInfo)(nil), "models.BuildRuntimeInfo")
	proto.RegisterType((*LogResponse)(nil), "models.LogResponse")
	proto.RegisterType((*RepoAccount)(nil), "models.RepoAccount")
	proto.RegisterType((*BuildSummary)(nil), "models.BuildSummary")
	proto.RegisterType((*Summaries)(nil), "models.Summaries")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GuideOcelot service

type GuideOcelotClient interface {
	GetVCSCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
	SetVCSCreds(ctx context.Context, in *VCSCreds, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
	GetRepoCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*RepoCredWrapper, error)
	SetRepoCreds(ctx context.Context, in *RepoCreds, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
	GetAllCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*AllCredsWrapper, error)
	CheckConn(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
	BuildRuntime(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (*BuildRuntimeInfo, error)
	Logs(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (GuideOcelot_LogsClient, error)
	LastFewSummaries(ctx context.Context, in *RepoAccount, opts ...grpc.CallOption) (*Summaries, error)
}

type guideOcelotClient struct {
	cc *grpc.ClientConn
}

func NewGuideOcelotClient(cc *grpc.ClientConn) GuideOcelotClient {
	return &guideOcelotClient{cc}
}

func (c *guideOcelotClient) GetVCSCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*CredWrapper, error) {
	out := new(CredWrapper)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/GetVCSCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) SetVCSCreds(ctx context.Context, in *VCSCreds, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/SetVCSCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) GetRepoCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*RepoCredWrapper, error) {
	out := new(RepoCredWrapper)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/GetRepoCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) SetRepoCreds(ctx context.Context, in *RepoCreds, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/SetRepoCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) GetAllCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*AllCredsWrapper, error) {
	out := new(AllCredsWrapper)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/GetAllCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) CheckConn(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/CheckConn", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) BuildRuntime(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (*BuildRuntimeInfo, error) {
	out := new(BuildRuntimeInfo)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/BuildRuntime", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) Logs(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (GuideOcelot_LogsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_GuideOcelot_serviceDesc.Streams[0], c.cc, "/models.GuideOcelot/Logs", opts...)
	if err != nil {
		return nil, err
	}
	x := &guideOcelotLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GuideOcelot_LogsClient interface {
	Recv() (*LogResponse, error)
	grpc.ClientStream
}

type guideOcelotLogsClient struct {
	grpc.ClientStream
}

func (x *guideOcelotLogsClient) Recv() (*LogResponse, error) {
	m := new(LogResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *guideOcelotClient) LastFewSummaries(ctx context.Context, in *RepoAccount, opts ...grpc.CallOption) (*Summaries, error) {
	out := new(Summaries)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/LastFewSummaries", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for GuideOcelot service

type GuideOcelotServer interface {
	GetVCSCreds(context.Context, *google_protobuf.Empty) (*CredWrapper, error)
	SetVCSCreds(context.Context, *VCSCreds) (*google_protobuf.Empty, error)
	GetRepoCreds(context.Context, *google_protobuf.Empty) (*RepoCredWrapper, error)
	SetRepoCreds(context.Context, *RepoCreds) (*google_protobuf.Empty, error)
	GetAllCreds(context.Context, *google_protobuf.Empty) (*AllCredsWrapper, error)
	CheckConn(context.Context, *google_protobuf.Empty) (*google_protobuf.Empty, error)
	BuildRuntime(context.Context, *BuildQuery) (*BuildRuntimeInfo, error)
	Logs(*BuildQuery, GuideOcelot_LogsServer) error
	LastFewSummaries(context.Context, *RepoAccount) (*Summaries, error)
}

func RegisterGuideOcelotServer(s *grpc.Server, srv GuideOcelotServer) {
	s.RegisterService(&_GuideOcelot_serviceDesc, srv)
}

func _GuideOcelot_GetVCSCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).GetVCSCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/GetVCSCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).GetVCSCreds(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_SetVCSCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VCSCreds)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).SetVCSCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/SetVCSCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).SetVCSCreds(ctx, req.(*VCSCreds))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_GetRepoCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).GetRepoCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/GetRepoCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).GetRepoCreds(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_SetRepoCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepoCreds)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).SetRepoCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/SetRepoCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).SetRepoCreds(ctx, req.(*RepoCreds))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_GetAllCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).GetAllCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/GetAllCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).GetAllCreds(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_CheckConn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).CheckConn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/CheckConn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).CheckConn(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_BuildRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).BuildRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/BuildRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).BuildRuntime(ctx, req.(*BuildQuery))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_Logs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BuildQuery)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GuideOcelotServer).Logs(m, &guideOcelotLogsServer{stream})
}

type GuideOcelot_LogsServer interface {
	Send(*LogResponse) error
	grpc.ServerStream
}

type guideOcelotLogsServer struct {
	grpc.ServerStream
}

func (x *guideOcelotLogsServer) Send(m *LogResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _GuideOcelot_LastFewSummaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepoAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).LastFewSummaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/LastFewSummaries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).LastFewSummaries(ctx, req.(*RepoAccount))
	}
	return interceptor(ctx, in, info, handler)
}

var _GuideOcelot_serviceDesc = grpc.ServiceDesc{
	ServiceName: "models.GuideOcelot",
	HandlerType: (*GuideOcelotServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVCSCreds",
			Handler:    _GuideOcelot_GetVCSCreds_Handler,
		},
		{
			MethodName: "SetVCSCreds",
			Handler:    _GuideOcelot_SetVCSCreds_Handler,
		},
		{
			MethodName: "GetRepoCreds",
			Handler:    _GuideOcelot_GetRepoCreds_Handler,
		},
		{
			MethodName: "SetRepoCreds",
			Handler:    _GuideOcelot_SetRepoCreds_Handler,
		},
		{
			MethodName: "GetAllCreds",
			Handler:    _GuideOcelot_GetAllCreds_Handler,
		},
		{
			MethodName: "CheckConn",
			Handler:    _GuideOcelot_CheckConn_Handler,
		},
		{
			MethodName: "BuildRuntime",
			Handler:    _GuideOcelot_BuildRuntime_Handler,
		},
		{
			MethodName: "LastFewSummaries",
			Handler:    _GuideOcelot_LastFewSummaries_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Logs",
			Handler:       _GuideOcelot_Logs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "guideocelot.proto",
}

func init() { proto.RegisterFile("guideocelot.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 795 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xdd, 0x4e, 0xdb, 0x48,
	0x14, 0x8e, 0x93, 0x10, 0xe2, 0xe3, 0x2c, 0x3f, 0x03, 0x62, 0x2d, 0xef, 0x6a, 0x17, 0x8d, 0x76,
	0x25, 0x54, 0xa9, 0x49, 0x01, 0x21, 0x21, 0xfa, 0xa3, 0xd2, 0xb4, 0x45, 0x48, 0xa1, 0x2d, 0x13,
	0x68, 0xaf, 0x1d, 0x67, 0x48, 0x2c, 0x6c, 0x8f, 0xe5, 0x19, 0x83, 0x72, 0xdb, 0xfb, 0x5e, 0x55,
	0xea, 0x8b, 0xf5, 0x01, 0x7a, 0xd3, 0xa7, 0xe8, 0x55, 0x35, 0xe3, 0xb1, 0xe3, 0x04, 0xa2, 0xaa,
	0x77, 0x73, 0xfe, 0x66, 0xbe, 0xf3, 0x9d, 0xf9, 0x0e, 0xac, 0x8f, 0x52, 0x7f, 0x48, 0x99, 0x47,
	0x03, 0x26, 0xda, 0x71, 0xc2, 0x04, 0x43, 0x8d, 0x90, 0x0d, 0x69, 0xc0, 0x9d, 0xbf, 0x46, 0x8c,
	0x8d, 0x02, 0xda, 0x51, 0xde, 0x41, 0x7a, 0xd5, 0xa1, 0x61, 0x2c, 0x26, 0x59, 0x92, 0xf3, 0xb7,
	0x0e, 0xba, 0xb1, 0xdf, 0x71, 0xa3, 0x88, 0x09, 0x57, 0xf8, 0x2c, 0xe2, 0x3a, 0xfa, 0xef, 0x7c,
	0xa9, 0xf0, 0x43, 0xca, 0x85, 0x1b, 0xc6, 0x59, 0x02, 0x9e, 0xc0, 0xea, 0x71, 0x10, 0x74, 0x13,
	0x3a, 0xe4, 0x1f, 0x12, 0x37, 0x8e, 0x69, 0x82, 0x0e, 0xc0, 0x4c, 0x68, 0xcc, 0x94, 0xcf, 0x36,
	0xb6, 0x8d, 0x1d, 0x6b, 0xef, 0xcf, 0x76, 0x06, 0xa5, 0x4d, 0x74, 0x40, 0xe7, 0x92, 0x69, 0x26,
	0xea, 0x40, 0xf3, 0xc6, 0xe3, 0x59, 0x55, 0x4d, 0x55, 0x6d, 0xe4, 0x55, 0xe5, 0x8a, 0x22, 0x09,
	0xef, 0x82, 0x55, 0x0a, 0x20, 0x0c, 0xb5, 0x1b, 0x8f, 0xdb, 0xd5, 0xed, 0xda, 0x8e, 0xb5, 0xb7,
	0x96, 0x97, 0xbe, 0xef, 0xf6, 0x55, 0x36, 0x91, 0x41, 0xfc, 0xc5, 0x80, 0x66, 0xee, 0x41, 0x0e,
	0x34, 0xbd, 0xc0, 0xa7, 0x91, 0x38, 0x1d, 0x2a, 0x98, 0x26, 0x29, 0x6c, 0x84, 0xa1, 0x95, 0x9d,
	0xfb, 0xd4, 0x4b, 0xa8, 0xb0, 0xab, 0x2a, 0x3e, 0xe3, 0x93, 0xf5, 0x82, 0x5d, 0xd3, 0xe8, 0x92,
	0xf4, 0x14, 0x60, 0x93, 0x14, 0xb6, 0x8c, 0xb9, 0x9e, 0x27, 0xde, 0xb8, 0x21, 0xb5, 0xeb, 0x59,
	0x2c, 0xb7, 0x11, 0x82, 0xba, 0x98, 0xc4, 0xd4, 0x5e, 0x52, 0x7e, 0x75, 0xc6, 0x87, 0xb0, 0x3a,
	0x47, 0x0d, 0xfa, 0x1f, 0xea, 0x92, 0x1c, 0xbb, 0xa6, 0x1a, 0x5a, 0x9f, 0x67, 0x90, 0x13, 0x15,
	0xc6, 0x9f, 0x0c, 0x30, 0x0b, 0x9f, 0x7c, 0x37, 0xe5, 0x34, 0x89, 0xe4, 0xbb, 0xba, 0xa7, 0xdc,
	0x96, 0xb1, 0xd8, 0xe5, 0xfc, 0x96, 0x25, 0x43, 0xdd, 0x4f, 0x61, 0x23, 0x1b, 0x96, 0xe5, 0x6d,
	0x97, 0x49, 0xa0, 0x5b, 0xc9, 0xcd, 0xdf, 0xee, 0xe4, 0x08, 0xe0, 0x45, 0xea, 0x07, 0xc3, 0xf3,
	0x94, 0x26, 0x13, 0x99, 0x31, 0x76, 0xf9, 0x58, 0x63, 0x51, 0x67, 0xf9, 0xd6, 0x40, 0x66, 0x9c,
	0x66, 0x30, 0x6a, 0x24, 0x37, 0x31, 0x81, 0x35, 0x55, 0x4b, 0xd2, 0x48, 0xfe, 0xb3, 0xd3, 0xe8,
	0x8a, 0xc9, 0x1b, 0x86, 0x2c, 0xca, 0xba, 0x69, 0x12, 0x75, 0x46, 0x2b, 0x50, 0xf5, 0x63, 0xdd,
	0x43, 0xd5, 0x8f, 0x25, 0xc6, 0x51, 0x12, 0x7b, 0xef, 0x58, 0x22, 0xf2, 0x49, 0xe4, 0x36, 0x7e,
	0x08, 0x56, 0x8f, 0x8d, 0x08, 0xe5, 0x31, 0x8b, 0x38, 0x45, 0xff, 0x00, 0xb0, 0x54, 0xc4, 0xa9,
	0xe8, 0xf9, 0x51, 0x4e, 0x51, 0xc9, 0x83, 0xcf, 0xc1, 0x92, 0x6c, 0x1e, 0x7b, 0x1e, 0x4b, 0x23,
	0x21, 0x5f, 0x57, 0x43, 0xd0, 0xf8, 0xe5, 0x59, 0xe2, 0x77, 0xb3, 0xb0, 0x86, 0x90, 0x9b, 0x68,
	0x13, 0x96, 0x02, 0x3f, 0xf4, 0x33, 0x10, 0x4b, 0x24, 0x33, 0xf0, 0x0f, 0x03, 0x5a, 0xaa, 0xad,
	0x7e, 0x1a, 0x86, 0xee, 0x02, 0x52, 0xb6, 0xa0, 0x71, 0xe5, 0xfa, 0x01, 0xcd, 0x38, 0x69, 0x12,
	0x6d, 0xa1, 0x43, 0x30, 0x15, 0x3b, 0x17, 0x7e, 0x48, 0xb5, 0x2c, 0x9c, 0x76, 0x26, 0xca, 0x76,
	0x2e, 0xca, 0xf6, 0x45, 0x2e, 0x4a, 0x32, 0x4d, 0x2e, 0xc3, 0xac, 0xcf, 0xc2, 0xfc, 0x0f, 0xfe,
	0x50, 0x69, 0x2f, 0xd3, 0x44, 0x89, 0x5d, 0xcd, 0xcf, 0x20, 0xb3, 0xce, 0xa2, 0xf5, 0x46, 0xa9,
	0xf5, 0x2d, 0x68, 0x0c, 0x12, 0x37, 0xf2, 0xc6, 0xf6, 0xb2, 0xf2, 0x6a, 0xab, 0x3c, 0xd2, 0xe6,
	0xec, 0x48, 0x0f, 0xc0, 0xcc, 0xda, 0xf6, 0x29, 0x47, 0x3b, 0x50, 0xe7, 0x69, 0x28, 0x97, 0x82,
	0xfc, 0xd2, 0x9b, 0xf9, 0x97, 0x2e, 0x93, 0x43, 0x54, 0xc6, 0xde, 0xb7, 0x3a, 0x58, 0x27, 0x72,
	0xa1, 0xbd, 0x55, 0x0b, 0x0d, 0x9d, 0x81, 0x75, 0x42, 0x45, 0x21, 0xdd, 0xad, 0x3b, 0x14, 0xbc,
	0x92, 0x2b, 0xcd, 0xb9, 0x6f, 0x63, 0xe0, 0xf5, 0x8f, 0x5f, 0xbf, 0x7f, 0xae, 0x5a, 0xc8, 0xec,
	0xdc, 0xec, 0x76, 0x3c, 0x55, 0x7f, 0x06, 0x56, 0xbf, 0x74, 0xdd, 0x9d, 0x6d, 0xe1, 0x2c, 0x78,
	0x00, 0x6f, 0xaa, 0xbb, 0x56, 0xf0, 0xf4, 0xae, 0x23, 0xe3, 0x01, 0x3a, 0x86, 0xd6, 0x09, 0x15,
	0x53, 0x15, 0x2e, 0x82, 0xb7, 0x68, 0x0d, 0xe2, 0x0a, 0x7a, 0x0c, 0xad, 0x7e, 0xf9, 0x8a, 0xbb,
	0x7a, 0x5f, 0x88, 0xa9, 0x82, 0x9e, 0x2b, 0x76, 0xf2, 0x3d, 0xfc, 0xeb, 0xe7, 0xe7, 0x36, 0x36,
	0xae, 0xa0, 0xa7, 0x60, 0x76, 0xc7, 0xd4, 0xbb, 0xee, 0xb2, 0x28, 0x5a, 0x58, 0xbf, 0x18, 0xc0,
	0x33, 0xfd, 0xc3, 0xb5, 0x70, 0x11, 0x9a, 0x19, 0xad, 0x5a, 0x05, 0x8e, 0x3d, 0xe3, 0x2b, 0x49,
	0x1c, 0x57, 0xd0, 0x3e, 0xd4, 0x7b, 0x6c, 0xc4, 0xef, 0xad, 0x2b, 0x66, 0x5a, 0x92, 0x31, 0xae,
	0x3c, 0x32, 0xd0, 0x13, 0x58, 0xeb, 0xb9, 0x5c, 0xbc, 0xa6, 0xb7, 0xd3, 0x1f, 0xb6, 0x51, 0xa6,
	0x4d, 0x8b, 0xd8, 0x29, 0xb8, 0x2c, 0xf2, 0x70, 0x65, 0xd0, 0x50, 0x4d, 0xec, 0xff, 0x0c, 0x00,
	0x00, 0xff, 0xff, 0x9f, 0x22, 0x1e, 0xde, 0x38, 0x07, 0x00, 0x00,
}
