// Copyright 2020 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.3
// source: api.proto

package results_go_proto

import (
	context "context"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Results_CreateResult_FullMethodName          = "/tekton.results.v1alpha2.Results/CreateResult"
	Results_UpdateResult_FullMethodName          = "/tekton.results.v1alpha2.Results/UpdateResult"
	Results_GetResult_FullMethodName             = "/tekton.results.v1alpha2.Results/GetResult"
	Results_DeleteResult_FullMethodName          = "/tekton.results.v1alpha2.Results/DeleteResult"
	Results_ListResults_FullMethodName           = "/tekton.results.v1alpha2.Results/ListResults"
	Results_CreateRecord_FullMethodName          = "/tekton.results.v1alpha2.Results/CreateRecord"
	Results_UpdateRecord_FullMethodName          = "/tekton.results.v1alpha2.Results/UpdateRecord"
	Results_GetRecord_FullMethodName             = "/tekton.results.v1alpha2.Results/GetRecord"
	Results_ListRecords_FullMethodName           = "/tekton.results.v1alpha2.Results/ListRecords"
	Results_DeleteRecord_FullMethodName          = "/tekton.results.v1alpha2.Results/DeleteRecord"
	Results_GetResultSummary_FullMethodName      = "/tekton.results.v1alpha2.Results/GetResultSummary"
	Results_GetResultsListSummary_FullMethodName = "/tekton.results.v1alpha2.Results/GetResultsListSummary"
)

// ResultsClient is the client API for Results service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ResultsClient interface {
	CreateResult(ctx context.Context, in *CreateResultRequest, opts ...grpc.CallOption) (*Result, error)
	UpdateResult(ctx context.Context, in *UpdateResultRequest, opts ...grpc.CallOption) (*Result, error)
	GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*Result, error)
	DeleteResult(ctx context.Context, in *DeleteResultRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListResults(ctx context.Context, in *ListResultsRequest, opts ...grpc.CallOption) (*ListResultsResponse, error)
	CreateRecord(ctx context.Context, in *CreateRecordRequest, opts ...grpc.CallOption) (*Record, error)
	UpdateRecord(ctx context.Context, in *UpdateRecordRequest, opts ...grpc.CallOption) (*Record, error)
	GetRecord(ctx context.Context, in *GetRecordRequest, opts ...grpc.CallOption) (*Record, error)
	ListRecords(ctx context.Context, in *ListRecordsRequest, opts ...grpc.CallOption) (*ListRecordsResponse, error)
	DeleteRecord(ctx context.Context, in *DeleteRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetResultSummary(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*Summary, error)
	GetResultsListSummary(ctx context.Context, in *ResultListSummaryRequest, opts ...grpc.CallOption) (*Summary, error)
}

type resultsClient struct {
	cc grpc.ClientConnInterface
}

func NewResultsClient(cc grpc.ClientConnInterface) ResultsClient {
	return &resultsClient{cc}
}

func (c *resultsClient) CreateResult(ctx context.Context, in *CreateResultRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, Results_CreateResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) UpdateResult(ctx context.Context, in *UpdateResultRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, Results_UpdateResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, Results_GetResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) DeleteResult(ctx context.Context, in *DeleteResultRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Results_DeleteResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) ListResults(ctx context.Context, in *ListResultsRequest, opts ...grpc.CallOption) (*ListResultsResponse, error) {
	out := new(ListResultsResponse)
	err := c.cc.Invoke(ctx, Results_ListResults_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) CreateRecord(ctx context.Context, in *CreateRecordRequest, opts ...grpc.CallOption) (*Record, error) {
	out := new(Record)
	err := c.cc.Invoke(ctx, Results_CreateRecord_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) UpdateRecord(ctx context.Context, in *UpdateRecordRequest, opts ...grpc.CallOption) (*Record, error) {
	out := new(Record)
	err := c.cc.Invoke(ctx, Results_UpdateRecord_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetRecord(ctx context.Context, in *GetRecordRequest, opts ...grpc.CallOption) (*Record, error) {
	out := new(Record)
	err := c.cc.Invoke(ctx, Results_GetRecord_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) ListRecords(ctx context.Context, in *ListRecordsRequest, opts ...grpc.CallOption) (*ListRecordsResponse, error) {
	out := new(ListRecordsResponse)
	err := c.cc.Invoke(ctx, Results_ListRecords_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) DeleteRecord(ctx context.Context, in *DeleteRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Results_DeleteRecord_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetResultSummary(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*Summary, error) {
	out := new(Summary)
	err := c.cc.Invoke(ctx, Results_GetResultSummary_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetResultsListSummary(ctx context.Context, in *ResultListSummaryRequest, opts ...grpc.CallOption) (*Summary, error) {
	out := new(Summary)
	err := c.cc.Invoke(ctx, Results_GetResultsListSummary_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ResultsServer is the server API for Results service.
// All implementations must embed UnimplementedResultsServer
// for forward compatibility
type ResultsServer interface {
	CreateResult(context.Context, *CreateResultRequest) (*Result, error)
	UpdateResult(context.Context, *UpdateResultRequest) (*Result, error)
	GetResult(context.Context, *GetResultRequest) (*Result, error)
	DeleteResult(context.Context, *DeleteResultRequest) (*emptypb.Empty, error)
	ListResults(context.Context, *ListResultsRequest) (*ListResultsResponse, error)
	CreateRecord(context.Context, *CreateRecordRequest) (*Record, error)
	UpdateRecord(context.Context, *UpdateRecordRequest) (*Record, error)
	GetRecord(context.Context, *GetRecordRequest) (*Record, error)
	ListRecords(context.Context, *ListRecordsRequest) (*ListRecordsResponse, error)
	DeleteRecord(context.Context, *DeleteRecordRequest) (*emptypb.Empty, error)
	GetResultSummary(context.Context, *GetResultRequest) (*Summary, error)
	GetResultsListSummary(context.Context, *ResultListSummaryRequest) (*Summary, error)
	mustEmbedUnimplementedResultsServer()
}

// UnimplementedResultsServer must be embedded to have forward compatible implementations.
type UnimplementedResultsServer struct {
}

func (UnimplementedResultsServer) CreateResult(context.Context, *CreateResultRequest) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateResult not implemented")
}
func (UnimplementedResultsServer) UpdateResult(context.Context, *UpdateResultRequest) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateResult not implemented")
}
func (UnimplementedResultsServer) GetResult(context.Context, *GetResultRequest) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResult not implemented")
}
func (UnimplementedResultsServer) DeleteResult(context.Context, *DeleteResultRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteResult not implemented")
}
func (UnimplementedResultsServer) ListResults(context.Context, *ListResultsRequest) (*ListResultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListResults not implemented")
}
func (UnimplementedResultsServer) CreateRecord(context.Context, *CreateRecordRequest) (*Record, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRecord not implemented")
}
func (UnimplementedResultsServer) UpdateRecord(context.Context, *UpdateRecordRequest) (*Record, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRecord not implemented")
}
func (UnimplementedResultsServer) GetRecord(context.Context, *GetRecordRequest) (*Record, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRecord not implemented")
}
func (UnimplementedResultsServer) ListRecords(context.Context, *ListRecordsRequest) (*ListRecordsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRecords not implemented")
}
func (UnimplementedResultsServer) DeleteRecord(context.Context, *DeleteRecordRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRecord not implemented")
}
func (UnimplementedResultsServer) GetResultSummary(context.Context, *GetResultRequest) (*Summary, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResultSummary not implemented")
}
func (UnimplementedResultsServer) GetResultsListSummary(context.Context, *ResultListSummaryRequest) (*Summary, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResultsListSummary not implemented")
}
func (UnimplementedResultsServer) mustEmbedUnimplementedResultsServer() {}

// UnsafeResultsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ResultsServer will
// result in compilation errors.
type UnsafeResultsServer interface {
	mustEmbedUnimplementedResultsServer()
}

func RegisterResultsServer(s grpc.ServiceRegistrar, srv ResultsServer) {
	s.RegisterService(&Results_ServiceDesc, srv)
}

func _Results_CreateResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).CreateResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_CreateResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).CreateResult(ctx, req.(*CreateResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_UpdateResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).UpdateResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_UpdateResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).UpdateResult(ctx, req.(*UpdateResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_GetResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).GetResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_GetResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).GetResult(ctx, req.(*GetResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_DeleteResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).DeleteResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_DeleteResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).DeleteResult(ctx, req.(*DeleteResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_ListResults_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListResultsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).ListResults(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_ListResults_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).ListResults(ctx, req.(*ListResultsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_CreateRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).CreateRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_CreateRecord_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).CreateRecord(ctx, req.(*CreateRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_UpdateRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).UpdateRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_UpdateRecord_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).UpdateRecord(ctx, req.(*UpdateRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_GetRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).GetRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_GetRecord_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).GetRecord(ctx, req.(*GetRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_ListRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).ListRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_ListRecords_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).ListRecords(ctx, req.(*ListRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_DeleteRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).DeleteRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_DeleteRecord_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).DeleteRecord(ctx, req.(*DeleteRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_GetResultSummary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).GetResultSummary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_GetResultSummary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).GetResultSummary(ctx, req.(*GetResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_GetResultsListSummary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResultListSummaryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).GetResultsListSummary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Results_GetResultsListSummary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).GetResultsListSummary(ctx, req.(*ResultListSummaryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Results_ServiceDesc is the grpc.ServiceDesc for Results service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Results_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tekton.results.v1alpha2.Results",
	HandlerType: (*ResultsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateResult",
			Handler:    _Results_CreateResult_Handler,
		},
		{
			MethodName: "UpdateResult",
			Handler:    _Results_UpdateResult_Handler,
		},
		{
			MethodName: "GetResult",
			Handler:    _Results_GetResult_Handler,
		},
		{
			MethodName: "DeleteResult",
			Handler:    _Results_DeleteResult_Handler,
		},
		{
			MethodName: "ListResults",
			Handler:    _Results_ListResults_Handler,
		},
		{
			MethodName: "CreateRecord",
			Handler:    _Results_CreateRecord_Handler,
		},
		{
			MethodName: "UpdateRecord",
			Handler:    _Results_UpdateRecord_Handler,
		},
		{
			MethodName: "GetRecord",
			Handler:    _Results_GetRecord_Handler,
		},
		{
			MethodName: "ListRecords",
			Handler:    _Results_ListRecords_Handler,
		},
		{
			MethodName: "DeleteRecord",
			Handler:    _Results_DeleteRecord_Handler,
		},
		{
			MethodName: "GetResultSummary",
			Handler:    _Results_GetResultSummary_Handler,
		},
		{
			MethodName: "GetResultsListSummary",
			Handler:    _Results_GetResultsListSummary_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}

const (
	Logs_GetLog_FullMethodName    = "/tekton.results.v1alpha2.Logs/GetLog"
	Logs_ListLogs_FullMethodName  = "/tekton.results.v1alpha2.Logs/ListLogs"
	Logs_UpdateLog_FullMethodName = "/tekton.results.v1alpha2.Logs/UpdateLog"
	Logs_DeleteLog_FullMethodName = "/tekton.results.v1alpha2.Logs/DeleteLog"
)

// LogsClient is the client API for Logs service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogsClient interface {
	GetLog(ctx context.Context, in *GetLogRequest, opts ...grpc.CallOption) (Logs_GetLogClient, error)
	ListLogs(ctx context.Context, in *ListRecordsRequest, opts ...grpc.CallOption) (*ListRecordsResponse, error)
	UpdateLog(ctx context.Context, opts ...grpc.CallOption) (Logs_UpdateLogClient, error)
	DeleteLog(ctx context.Context, in *DeleteLogRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type logsClient struct {
	cc grpc.ClientConnInterface
}

func NewLogsClient(cc grpc.ClientConnInterface) LogsClient {
	return &logsClient{cc}
}

func (c *logsClient) GetLog(ctx context.Context, in *GetLogRequest, opts ...grpc.CallOption) (Logs_GetLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Logs_ServiceDesc.Streams[0], Logs_GetLog_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &logsGetLogClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Logs_GetLogClient interface {
	Recv() (*httpbody.HttpBody, error)
	grpc.ClientStream
}

type logsGetLogClient struct {
	grpc.ClientStream
}

func (x *logsGetLogClient) Recv() (*httpbody.HttpBody, error) {
	m := new(httpbody.HttpBody)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *logsClient) ListLogs(ctx context.Context, in *ListRecordsRequest, opts ...grpc.CallOption) (*ListRecordsResponse, error) {
	out := new(ListRecordsResponse)
	err := c.cc.Invoke(ctx, Logs_ListLogs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logsClient) UpdateLog(ctx context.Context, opts ...grpc.CallOption) (Logs_UpdateLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Logs_ServiceDesc.Streams[1], Logs_UpdateLog_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &logsUpdateLogClient{stream}
	return x, nil
}

type Logs_UpdateLogClient interface {
	Send(*Log) error
	CloseAndRecv() (*LogSummary, error)
	grpc.ClientStream
}

type logsUpdateLogClient struct {
	grpc.ClientStream
}

func (x *logsUpdateLogClient) Send(m *Log) error {
	return x.ClientStream.SendMsg(m)
}

func (x *logsUpdateLogClient) CloseAndRecv() (*LogSummary, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(LogSummary)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *logsClient) DeleteLog(ctx context.Context, in *DeleteLogRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Logs_DeleteLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogsServer is the server API for Logs service.
// All implementations must embed UnimplementedLogsServer
// for forward compatibility
type LogsServer interface {
	GetLog(*GetLogRequest, Logs_GetLogServer) error
	ListLogs(context.Context, *ListRecordsRequest) (*ListRecordsResponse, error)
	UpdateLog(Logs_UpdateLogServer) error
	DeleteLog(context.Context, *DeleteLogRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedLogsServer()
}

// UnimplementedLogsServer must be embedded to have forward compatible implementations.
type UnimplementedLogsServer struct {
}

func (UnimplementedLogsServer) GetLog(*GetLogRequest, Logs_GetLogServer) error {
	return status.Errorf(codes.Unimplemented, "method GetLog not implemented")
}
func (UnimplementedLogsServer) ListLogs(context.Context, *ListRecordsRequest) (*ListRecordsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLogs not implemented")
}
func (UnimplementedLogsServer) UpdateLog(Logs_UpdateLogServer) error {
	return status.Errorf(codes.Unimplemented, "method UpdateLog not implemented")
}
func (UnimplementedLogsServer) DeleteLog(context.Context, *DeleteLogRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLog not implemented")
}
func (UnimplementedLogsServer) mustEmbedUnimplementedLogsServer() {}

// UnsafeLogsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogsServer will
// result in compilation errors.
type UnsafeLogsServer interface {
	mustEmbedUnimplementedLogsServer()
}

func RegisterLogsServer(s grpc.ServiceRegistrar, srv LogsServer) {
	s.RegisterService(&Logs_ServiceDesc, srv)
}

func _Logs_GetLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetLogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LogsServer).GetLog(m, &logsGetLogServer{stream})
}

type Logs_GetLogServer interface {
	Send(*httpbody.HttpBody) error
	grpc.ServerStream
}

type logsGetLogServer struct {
	grpc.ServerStream
}

func (x *logsGetLogServer) Send(m *httpbody.HttpBody) error {
	return x.ServerStream.SendMsg(m)
}

func _Logs_ListLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogsServer).ListLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Logs_ListLogs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogsServer).ListLogs(ctx, req.(*ListRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Logs_UpdateLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LogsServer).UpdateLog(&logsUpdateLogServer{stream})
}

type Logs_UpdateLogServer interface {
	SendAndClose(*LogSummary) error
	Recv() (*Log, error)
	grpc.ServerStream
}

type logsUpdateLogServer struct {
	grpc.ServerStream
}

func (x *logsUpdateLogServer) SendAndClose(m *LogSummary) error {
	return x.ServerStream.SendMsg(m)
}

func (x *logsUpdateLogServer) Recv() (*Log, error) {
	m := new(Log)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Logs_DeleteLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogsServer).DeleteLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Logs_DeleteLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogsServer).DeleteLog(ctx, req.(*DeleteLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Logs_ServiceDesc is the grpc.ServiceDesc for Logs service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Logs_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tekton.results.v1alpha2.Logs",
	HandlerType: (*LogsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListLogs",
			Handler:    _Logs_ListLogs_Handler,
		},
		{
			MethodName: "DeleteLog",
			Handler:    _Logs_DeleteLog_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetLog",
			Handler:       _Logs_GetLog_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "UpdateLog",
			Handler:       _Logs_UpdateLog_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "api.proto",
}
