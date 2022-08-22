// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.4
// source: api.proto

package results_go_proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

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
	GetLog(ctx context.Context, in *GetLogRequest, opts ...grpc.CallOption) (Results_GetLogClient, error)
	ListLogs(ctx context.Context, in *ListLogsRequest, opts ...grpc.CallOption) (*ListLogsResponse, error)
	UpdateLog(ctx context.Context, opts ...grpc.CallOption) (Results_UpdateLogClient, error)
	DeleteLog(ctx context.Context, in *DeleteLogRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type resultsClient struct {
	cc grpc.ClientConnInterface
}

func NewResultsClient(cc grpc.ClientConnInterface) ResultsClient {
	return &resultsClient{cc}
}

func (c *resultsClient) CreateResult(ctx context.Context, in *CreateResultRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/CreateResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) UpdateResult(ctx context.Context, in *UpdateResultRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/UpdateResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/GetResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) DeleteResult(ctx context.Context, in *DeleteResultRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/DeleteResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) ListResults(ctx context.Context, in *ListResultsRequest, opts ...grpc.CallOption) (*ListResultsResponse, error) {
	out := new(ListResultsResponse)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/ListResults", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) CreateRecord(ctx context.Context, in *CreateRecordRequest, opts ...grpc.CallOption) (*Record, error) {
	out := new(Record)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/CreateRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) UpdateRecord(ctx context.Context, in *UpdateRecordRequest, opts ...grpc.CallOption) (*Record, error) {
	out := new(Record)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/UpdateRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetRecord(ctx context.Context, in *GetRecordRequest, opts ...grpc.CallOption) (*Record, error) {
	out := new(Record)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/GetRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) ListRecords(ctx context.Context, in *ListRecordsRequest, opts ...grpc.CallOption) (*ListRecordsResponse, error) {
	out := new(ListRecordsResponse)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/ListRecords", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) DeleteRecord(ctx context.Context, in *DeleteRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/DeleteRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) GetLog(ctx context.Context, in *GetLogRequest, opts ...grpc.CallOption) (Results_GetLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Results_ServiceDesc.Streams[0], "/tekton.results.v1alpha2.Results/GetLog", opts...)
	if err != nil {
		return nil, err
	}
	x := &resultsGetLogClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Results_GetLogClient interface {
	Recv() (*Log, error)
	grpc.ClientStream
}

type resultsGetLogClient struct {
	grpc.ClientStream
}

func (x *resultsGetLogClient) Recv() (*Log, error) {
	m := new(Log)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *resultsClient) ListLogs(ctx context.Context, in *ListLogsRequest, opts ...grpc.CallOption) (*ListLogsResponse, error) {
	out := new(ListLogsResponse)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/ListLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resultsClient) UpdateLog(ctx context.Context, opts ...grpc.CallOption) (Results_UpdateLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Results_ServiceDesc.Streams[1], "/tekton.results.v1alpha2.Results/UpdateLog", opts...)
	if err != nil {
		return nil, err
	}
	x := &resultsUpdateLogClient{stream}
	return x, nil
}

type Results_UpdateLogClient interface {
	Send(*Log) error
	CloseAndRecv() (*LogSummary, error)
	grpc.ClientStream
}

type resultsUpdateLogClient struct {
	grpc.ClientStream
}

func (x *resultsUpdateLogClient) Send(m *Log) error {
	return x.ClientStream.SendMsg(m)
}

func (x *resultsUpdateLogClient) CloseAndRecv() (*LogSummary, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(LogSummary)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *resultsClient) DeleteLog(ctx context.Context, in *DeleteLogRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/tekton.results.v1alpha2.Results/DeleteLog", in, out, opts...)
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
	GetLog(*GetLogRequest, Results_GetLogServer) error
	ListLogs(context.Context, *ListLogsRequest) (*ListLogsResponse, error)
	UpdateLog(Results_UpdateLogServer) error
	DeleteLog(context.Context, *DeleteLogRequest) (*emptypb.Empty, error)
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
func (UnimplementedResultsServer) GetLog(*GetLogRequest, Results_GetLogServer) error {
	return status.Errorf(codes.Unimplemented, "method GetLog not implemented")
}
func (UnimplementedResultsServer) ListLogs(context.Context, *ListLogsRequest) (*ListLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLogs not implemented")
}
func (UnimplementedResultsServer) UpdateLog(Results_UpdateLogServer) error {
	return status.Errorf(codes.Unimplemented, "method UpdateLog not implemented")
}
func (UnimplementedResultsServer) DeleteLog(context.Context, *DeleteLogRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLog not implemented")
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
		FullMethod: "/tekton.results.v1alpha2.Results/CreateResult",
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
		FullMethod: "/tekton.results.v1alpha2.Results/UpdateResult",
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
		FullMethod: "/tekton.results.v1alpha2.Results/GetResult",
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
		FullMethod: "/tekton.results.v1alpha2.Results/DeleteResult",
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
		FullMethod: "/tekton.results.v1alpha2.Results/ListResults",
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
		FullMethod: "/tekton.results.v1alpha2.Results/CreateRecord",
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
		FullMethod: "/tekton.results.v1alpha2.Results/UpdateRecord",
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
		FullMethod: "/tekton.results.v1alpha2.Results/GetRecord",
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
		FullMethod: "/tekton.results.v1alpha2.Results/ListRecords",
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
		FullMethod: "/tekton.results.v1alpha2.Results/DeleteRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).DeleteRecord(ctx, req.(*DeleteRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_GetLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetLogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ResultsServer).GetLog(m, &resultsGetLogServer{stream})
}

type Results_GetLogServer interface {
	Send(*Log) error
	grpc.ServerStream
}

type resultsGetLogServer struct {
	grpc.ServerStream
}

func (x *resultsGetLogServer) Send(m *Log) error {
	return x.ServerStream.SendMsg(m)
}

func _Results_ListLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).ListLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tekton.results.v1alpha2.Results/ListLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).ListLogs(ctx, req.(*ListLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Results_UpdateLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ResultsServer).UpdateLog(&resultsUpdateLogServer{stream})
}

type Results_UpdateLogServer interface {
	SendAndClose(*LogSummary) error
	Recv() (*Log, error)
	grpc.ServerStream
}

type resultsUpdateLogServer struct {
	grpc.ServerStream
}

func (x *resultsUpdateLogServer) SendAndClose(m *LogSummary) error {
	return x.ServerStream.SendMsg(m)
}

func (x *resultsUpdateLogServer) Recv() (*Log, error) {
	m := new(Log)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Results_DeleteLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResultsServer).DeleteLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tekton.results.v1alpha2.Results/DeleteLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResultsServer).DeleteLog(ctx, req.(*DeleteLogRequest))
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
			MethodName: "ListLogs",
			Handler:    _Results_ListLogs_Handler,
		},
		{
			MethodName: "DeleteLog",
			Handler:    _Results_DeleteLog_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetLog",
			Handler:       _Results_GetLog_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "UpdateLog",
			Handler:       _Results_UpdateLog_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "api.proto",
}
