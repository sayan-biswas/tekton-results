package results

import (
	"context"
	"fmt"
	"github.com/tektoncd/results/pkg/api/server/v1alpha2/record"
	"github.com/tektoncd/results/pkg/watcher/convert"
	"github.com/tektoncd/results/pkg/watcher/reconciler/annotation"
	pb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PutLog adds the given Object to the Results API.
// If the parent result is missing or the object is not yet associated with a
// result, one is created automatically.
func (c *Client) PutLog(ctx context.Context, o Object, opts ...grpc.CallOption) (*pb.Result, *pb.Record, error) {
	result, err := c.ensureResult(ctx, o, opts...)
	if err != nil {
		return nil, nil, err
	}
	logRecord, err := c.createLogRecord(ctx, result.GetName(), o, opts...)
	if err != nil {
		return nil, nil, err
	}
	return result, logRecord, nil
}

// createLogRecord creates a record for logs.
func (c *Client) createLogRecord(ctx context.Context, parent string, o Object, opts ...grpc.CallOption) (*pb.Record, error) {
	name := logName(parent, o)
	rec, err := c.GetRecord(ctx, &pb.GetRecordRequest{Name: name}, opts...)
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}
	if rec != nil {
		return rec, nil
	}
	data, err := convert.ToLogProto(o, name)
	if err != nil {
		return nil, err
	}
	return c.CreateRecord(ctx, &pb.CreateRecordRequest{
		Parent: parent,
		Record: &pb.Record{
			Name: name,
			Data: data,
		},
	})
}

// logName gets the log name to use for the given object.
// The name is derived from a known Tekton annotation if available, else
// the object's name is used.
func logName(parent string, o Object) string {
	name, ok := o.GetAnnotations()[annotation.Log]
	if ok {
		return name
	}
	return record.FormatName(parent, defaultLogName(o))
}

// defaultLogName is the default Record name that should be used for a log record if one is not
// already associated to the Object.
func defaultLogName(o metav1.Object) string {
	return fmt.Sprintf("%s-log", o.GetUID())
}
