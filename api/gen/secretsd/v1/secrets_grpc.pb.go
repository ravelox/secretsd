// Code generated manually for bootstrap; replace with protoc-gen-go-grpc output.

package v1

import (
	"context"
	"google.golang.org/grpc"
)

type SecretsServer interface {
	Put(context.Context, *PutRequest) (*PutResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
}

type UnimplementedSecretsServer struct{}

func RegisterSecretsServer(s *grpc.Server, srv SecretsServer) {
	s.RegisterService(&Secrets_ServiceDesc, srv)
}

func _Secrets_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil {
		return srv.(SecretsServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/secretsd.v1.Secrets/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SecretsServer).Put(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Secrets_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil {
		return srv.(SecretsServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/secretsd.v1.Secrets/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SecretsServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Secrets_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "secretsd.v1.Secrets",
	HandlerType: (*SecretsServer)(nil),
	Methods: []grpc.MethodDesc{
		{ MethodName: "Put", Handler: _Secrets_Put_Handler },
		{ MethodName: "Get", Handler: _Secrets_Get_Handler },
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/secrets.proto",
}
