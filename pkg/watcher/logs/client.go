package logs

import (
	"context"
	"errors"
	v1alpha2pb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc"
	reflectionv1alpha "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"knative.dev/pkg/logging"
)

type Key struct{}

const (
	logsServiceName = "tekton.results.v1alpha2.Logs"
)

func WithContext(ctx context.Context, conn *grpc.ClientConn) (context.Context, error) {
	reflectionInfo, err := reflectionv1alpha.NewServerReflectionClient(conn).ServerReflectionInfo(ctx)
	if err != nil {
		return ctx, err
	}

	err = reflectionInfo.Send(&reflectionv1alpha.ServerReflectionRequest{
		MessageRequest: &reflectionv1alpha.ServerReflectionRequest_ListServices{
			ListServices: "*",
		},
	})
	if err != nil {
		return ctx, err
	}

	response, err := reflectionInfo.Recv()
	if err != nil {
		return ctx, err
	}
	for _, service := range response.GetListServicesResponse().GetService() {
		if service.Name == logsServiceName {
			return context.WithValue(ctx, Key{}, v1alpha2pb.NewLogsClient(conn)), nil
		}
	}
	return ctx, errors.New("logs service not enabled in server")
}

// Get extracts the Logs client from the context.
func Get(ctx context.Context) v1alpha2pb.LogsClient {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Info(
			"Unable to fetch Logs Client from context, either disabled from config or disabled from server side")
		return nil
	}
	return untyped.(v1alpha2pb.LogsClient)
}
