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

package server

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/results/pkg/internal/jsonutil"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2/record"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2/result"
	"github.com/tektoncd/results/pkg/server/db/pagination"
	"github.com/tektoncd/results/pkg/server/db/test"
	ppb "github.com/tektoncd/results/proto/pipeline/v1beta1/pipeline_go_proto"
	rpb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateRecord(t *testing.T) {
	srv, err := New(test.NewDB(t))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()
	res, err := srv.CreateResult(ctx, &rpb.CreateResultRequest{
		Parent: "foo",
		Result: &rpb.Result{
			Name: "foo/results/bar",
		},
	})
	if err != nil {
		t.Fatalf("CreateResult: %v", err)
	}

	req := &rpb.CreateRecordRequest{
		Parent: res.GetName(),
		Record: &rpb.Record{
			Name: record.FormatName(res.GetName(), "baz"),
			Data: &rpb.Any{
				Type:  "TaskRun",
				Value: jsonutil.AnyBytes(t, &v1beta1.TaskRun{ObjectMeta: v1.ObjectMeta{Name: "tacocat"}}),
			},
		},
	}
	t.Run("success", func(t *testing.T) {
		got, err := srv.CreateRecord(ctx, req)
		if err != nil {
			t.Fatalf("CreateRecord: %v", err)
		}
		want := proto.Clone(req.GetRecord()).(*rpb.Record)
		want.Id = fmt.Sprint(lastID)
		want.Uid = fmt.Sprint(lastID)
		want.CreatedTime = timestamppb.New(clock.Now())
		want.CreateTime = timestamppb.New(clock.Now())
		want.UpdatedTime = timestamppb.New(clock.Now())
		want.UpdateTime = timestamppb.New(clock.Now())
		want.Etag = mockEtag(lastID, clock.Now().UnixNano())

		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("-want, +got: %s", diff)
		}
	})

	// Errors
	for _, tc := range []struct {
		name string
		req  *rpb.CreateRecordRequest
		want codes.Code
	}{
		{
			name: "mismatched parent",
			req: &rpb.CreateRecordRequest{
				Parent: req.GetParent(),
				Record: &rpb.Record{
					Name: result.FormatName("foo", "baz"),
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "parent does not exist",
			req: &rpb.CreateRecordRequest{
				Parent: result.FormatName("foo", "doesnotexist"),
				Record: &rpb.Record{
					Name: record.FormatName(result.FormatName("foo", "doesnotexist"), "baz"),
				},
			},
			want: codes.NotFound,
		},
		{
			name: "missing name",
			req: &rpb.CreateRecordRequest{
				Parent: req.GetParent(),
				Record: &rpb.Record{
					Name: fmt.Sprintf("%s/results/", res.GetName()),
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "result used as name",
			req: &rpb.CreateRecordRequest{
				Parent: req.GetParent(),
				Record: &rpb.Record{
					Name: res.GetName(),
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "already exists",
			req:  req,
			want: codes.AlreadyExists,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := srv.CreateRecord(ctx, tc.req); status.Code(err) != tc.want {
				t.Fatalf("want: %v, got: %v - %+v", tc.want, status.Code(err), err)
			}
		})
	}
}

// TestCreateRecord_ConcurrentDelete simulates a concurrent deletion of a
// Result parent mocking the result name -> id conversion. This tricks the
// API Server into thinking the parent is valid during initial validation,
// but fails when writing the Record due to foreign key constraints.
func TestCreateRecord_ConcurrentDelete(t *testing.T) {
	res := "deleted"
	srv, err := New(
		test.NewDB(t),
		withGetResultID(func(context.Context, string, string) (string, error) {
			return res, nil
		}),
	)
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}

	ctx := context.Background()
	parent := result.FormatName("foo", res)
	rec, err := srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
		Parent: parent,
		Record: &rpb.Record{
			Name: record.FormatName(parent, "baz"),
		},
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("CreateRecord: %+v, %v", rec, err)
	}
}

func TestGetRecord(t *testing.T) {
	srv, err := New(test.NewDB(t))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()
	res, err := srv.CreateResult(ctx, &rpb.CreateResultRequest{
		Parent: "foo",
		Result: &rpb.Result{
			Name: "foo/results/bar",
		},
	})
	if err != nil {
		t.Fatalf("CreateResult: %v", err)
	}

	rec, err := srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
		Parent: res.GetName(),
		Record: &rpb.Record{
			Name: record.FormatName(res.GetName(), "baz"),
		},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		got, err := srv.GetRecord(ctx, &rpb.GetRecordRequest{Name: rec.GetName()})
		if err != nil {
			t.Fatalf("GetRecord: %v", err)
		}
		if diff := cmp.Diff(got, rec, protocmp.Transform()); diff != "" {
			t.Errorf("-want, +got: %s", diff)
		}
	})

	// Errors
	for _, tc := range []struct {
		name string
		req  *rpb.GetRecordRequest
		want codes.Code
	}{
		{
			name: "no name",
			req:  &rpb.GetRecordRequest{},
			want: codes.InvalidArgument,
		},
		{
			name: "invalid name",
			req:  &rpb.GetRecordRequest{Name: "a/results/doesnotexist"},
			want: codes.InvalidArgument,
		},
		{
			name: "not found",
			req:  &rpb.GetRecordRequest{Name: record.FormatName(res.GetName(), "doesnotexist")},
			want: codes.NotFound,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := srv.GetRecord(ctx, tc.req); status.Code(err) != tc.want {
				t.Fatalf("want: %v, got: %v - %+v", tc.want, status.Code(err), err)
			}
		})
	}
}

func TestListRecords(t *testing.T) {
	// Create a temporary database
	srv, err := New(test.NewDB(t))
	if err != nil {
		t.Fatalf("failed to setup db: %v", err)
	}
	ctx := context.Background()

	res, err := srv.CreateResult(ctx, &rpb.CreateResultRequest{
		Parent: "foo",
		Result: &rpb.Result{
			Name: "foo/results/bar",
		},
	})
	if err != nil {
		t.Fatalf("CreateResult: %v", err)
	}

	records := make([]*rpb.Record, 0, 6)
	// Create 3 TaskRun records
	for i := 0; i < 3; i++ {
		fakeClock.Advance(time.Second)
		rec, err := srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
			Parent: res.GetName(),
			Record: &rpb.Record{
				Name: fmt.Sprintf("%s/records/%d", res.GetName(), i),
				Data: &rpb.Any{
					Type: "TaskRun",
					Value: jsonutil.AnyBytes(t, &v1beta1.TaskRun{ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%d", i),
					}}),
				},
			},
		})
		if err != nil {
			t.Fatalf("could not create result: %v", err)
		}
		t.Logf("Created record: %+v", rec)
		records = append(records, rec)
	}

	// Create 3 PipelineRun records
	for i := 3; i < 6; i++ {
		fakeClock.Advance(time.Second)
		rec, err := srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
			Parent: res.GetName(),
			Record: &rpb.Record{
				Name: fmt.Sprintf("%s/records/%d", res.GetName(), i),
				Data: &rpb.Any{
					Type: "PipelineRun",
					Value: jsonutil.AnyBytes(t, &v1beta1.PipelineRun{ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%d", i),
					}}),
				},
			},
		})
		if err != nil {
			t.Fatalf("could not create result: %v", err)
		}
		t.Logf("Created record: %+v", rec)
		records = append(records, rec)
	}

	reversedRecords := make([]*rpb.Record, len(records))
	for i := len(reversedRecords); i > 0; i-- {
		reversedRecords[len(records)-i] = records[i-1]
	}

	tt := []struct {
		name   string
		req    *rpb.ListRecordsRequest
		want   *rpb.ListRecordsResponse
		status codes.Code
	}{
		{
			name: "all",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
			},
			want: &rpb.ListRecordsResponse{
				Records: records,
			},
		},
		{
			// TODO: We should return NOT_FOUND in the future.
			name: "missing parent",
			req: &rpb.ListRecordsRequest{
				Parent: "foo/results/baz",
			},
			want: &rpb.ListRecordsResponse{},
		},
		{
			name: "filter by record property",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
				Filter: `name == "foo/results/bar/records/0"`,
			},
			want: &rpb.ListRecordsResponse{
				Records: records[:1],
			},
		},
		{
			name: "filter by record data",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
				Filter: `data.metadata.name == "0"`,
			},
			want: &rpb.ListRecordsResponse{
				Records: records[:1],
			},
		},
		{
			name: "filter by record type",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
				Filter: `data_type == "TaskRun"`,
			},
			want: &rpb.ListRecordsResponse{
				Records: records[:3],
			},
		},
		{
			name: "filter by parent",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
				Filter: fmt.Sprintf(`name.startsWith("%s")`, res.GetName()),
			},
			want: &rpb.ListRecordsResponse{
				Records: records,
			},
		},
		// Pagination
		{
			name: "filter and page size",
			req: &rpb.ListRecordsRequest{
				Parent:   res.GetName(),
				Filter:   `data_type == "TaskRun"`,
				PageSize: 1,
			},
			want: &rpb.ListRecordsResponse{
				Records:       records[:1],
				NextPageToken: pageToken(t, records[1].GetId(), `data_type == "TaskRun"`),
			},
		},
		{
			name: "only page size",
			req: &rpb.ListRecordsRequest{
				Parent:   res.GetName(),
				PageSize: 1,
			},
			want: &rpb.ListRecordsResponse{
				Records:       records[:1],
				NextPageToken: pageToken(t, records[1].GetId(), ""),
			},
		},
		// Order By
		{
			name: "with order asc",
			req: &rpb.ListRecordsRequest{
				Parent:  res.GetName(),
				OrderBy: "created_time asc",
			},
			want: &rpb.ListRecordsResponse{
				Records: records,
			},
		},
		{
			name: "with order desc",
			req: &rpb.ListRecordsRequest{
				Parent:  res.GetName(),
				OrderBy: "created_time desc",
			},
			want: &rpb.ListRecordsResponse{
				Records: reversedRecords,
			},
		},
		{
			name: "with missing order",
			req: &rpb.ListRecordsRequest{
				Parent:  res.GetName(),
				OrderBy: "",
			},
			want: &rpb.ListRecordsResponse{
				Records: records,
			},
		},
		{
			name: "with default order",
			req: &rpb.ListRecordsRequest{
				Parent:  res.GetName(),
				OrderBy: "created_time",
			},
			want: &rpb.ListRecordsResponse{
				Records: records,
			},
		},

		// Errors
		{
			name: "unknown type",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
				Filter: `type(record.data) == tekton.pipeline.v1beta1.Unknown`,
			},
			status: codes.InvalidArgument,
		},
		{
			name: "unknown any field",
			req: &rpb.ListRecordsRequest{
				Parent: res.GetName(),
				Filter: `record.data.metadata.unknown == "tacocat"`,
			},
			status: codes.InvalidArgument,
		},
		{
			name: "invalid page size",
			req: &rpb.ListRecordsRequest{
				Parent:   res.GetName(),
				PageSize: -1,
			},
			status: codes.InvalidArgument,
		},
		{
			name: "malformed parent",
			req: &rpb.ListRecordsRequest{
				Parent: "unknown",
			},
			status: codes.InvalidArgument,
		},
		{
			name: "invalid order by clause",
			req: &rpb.ListRecordsRequest{
				Parent:  res.GetName(),
				OrderBy: "created_time desc asc",
			},
			status: codes.InvalidArgument,
		},
		{
			name: "invalid sort direction",
			req: &rpb.ListRecordsRequest{
				Parent:  res.GetName(),
				OrderBy: "created_time foo",
			},
			status: codes.InvalidArgument,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := srv.ListRecords(ctx, tc.req)
			if status.Code(err) != tc.status {
				t.Fatalf("want %v, got %v", tc.status, err)
			}

			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("-want, +got: %s", diff)
				if name, filter, err := pagination.DecodeToken(got.GetNextPageToken()); err == nil {
					t.Logf("Next (name, filter) = (%s, %s)", name, filter)
				}
			}
		})
	}
}

func TestUpdateRecord(t *testing.T) {
	srv, err := New(test.NewDB(t))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	ctx := context.Background()

	res, err := srv.CreateResult(ctx, &rpb.CreateResultRequest{
		Parent: "foo",
		Result: &rpb.Result{
			Name: result.FormatName("foo", "bar"),
		},
	})
	if err != nil {
		t.Fatalf("CreateResult(): %v", err)
	}

	tr := &ppb.TaskRun{
		Metadata: &ppb.ObjectMeta{
			Name: "taskrun",
		},
	}

	tt := []struct {
		name string
		// Starting Record to create.
		record *rpb.Record
		req    *rpb.UpdateRecordRequest
		// Expected update diff: expected Record should be merge of
		// record + diff.
		diff   *rpb.Record
		status codes.Code
	}{
		{
			name: "success",
			record: &rpb.Record{
				Name: record.FormatName(res.GetName(), "a"),
			},
			req: &rpb.UpdateRecordRequest{
				Etag: mockEtag(lastID+1, clock.Now().UnixNano()),
				Record: &rpb.Record{
					Name: record.FormatName(res.GetName(), "a"),
					Data: &rpb.Any{
						Value: jsonutil.AnyBytes(t, tr),
					},
				},
			},
			diff: &rpb.Record{
				Data: &rpb.Any{
					Value: jsonutil.AnyBytes(t, tr),
				},
			},
		},
		{
			name: "ignored fields",
			record: &rpb.Record{
				Name: record.FormatName(res.GetName(), "b"),
			},
			req: &rpb.UpdateRecordRequest{
				Record: &rpb.Record{
					Name: record.FormatName(res.GetName(), "b"),
					Id:   "ignored",
				},
			},
		},
		// Errors
		{
			name: "rename",
			req: &rpb.UpdateRecordRequest{
				Record: &rpb.Record{
					Name: record.FormatName(res.GetName(), "doesnotexist"),
					Data: &rpb.Any{
						Value: jsonutil.AnyBytes(t, tr),
					},
				},
			},
			status: codes.NotFound,
		},
		{
			name: "bad name",
			req: &rpb.UpdateRecordRequest{
				Record: &rpb.Record{
					Name: "tacocat",
					Data: &rpb.Any{
						Value: jsonutil.AnyBytes(t, tr),
					},
				},
			},
			status: codes.InvalidArgument,
		},
		{
			name: "etag mismatch",
			req: &rpb.UpdateRecordRequest{
				Etag: "invalid etag",
				Record: &rpb.Record{
					Name: record.FormatName(res.GetName(), "a"),
					Data: &rpb.Any{
						Value: jsonutil.AnyBytes(t, tr),
					},
				},
			},
			status: codes.FailedPrecondition,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var rec *rpb.Record
			if tc.record != nil {
				var err error
				rec, err = srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
					Parent: res.GetName(),
					Record: tc.record,
				})
				if err != nil {
					t.Fatalf("CreateRecord(): %v", err)
				}
			}

			fakeClock.Advance(time.Second)

			got, err := srv.UpdateRecord(ctx, tc.req)
			// if there is an error from UpdateRecord or expecting an error here,
			// compare the two errors.
			if err != nil || tc.status != codes.OK {
				if status.Code(err) == tc.status {
					return
				}
				t.Fatalf("UpdateRecord(%+v): %v", tc.req, err)
			}

			proto.Merge(rec, tc.diff)
			rec.UpdatedTime = timestamppb.New(clock.Now())
			rec.UpdateTime = timestamppb.New(clock.Now())
			rec.Etag = mockEtag(lastID, rec.UpdatedTime.AsTime().UnixNano())

			if diff := cmp.Diff(rec, got, protocmp.Transform()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestDeleteRecord(t *testing.T) {
	srv, err := New(test.NewDB(t))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()
	res, err := srv.CreateResult(ctx, &rpb.CreateResultRequest{
		Parent: "foo",
		Result: &rpb.Result{
			Name: "foo/results/bar",
		},
	})
	if err != nil {
		t.Fatalf("CreateResult: %v", err)
	}
	r, err := srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
		Parent: res.GetName(),
		Record: &rpb.Record{
			Name: record.FormatName(res.GetName(), "baz"),
		},
	})
	if err != nil {
		t.Fatalf("CreateRecord(): %v", err)
	}

	t.Run("success", func(t *testing.T) {
		// Delete inserted record
		if _, err := srv.DeleteRecord(ctx, &rpb.DeleteRecordRequest{Name: r.GetName()}); err != nil {
			t.Fatalf("could not delete record: %v", err)
		}
		// Check if the the record is deleted
		if r, err := srv.GetRecord(ctx, &rpb.GetRecordRequest{Name: r.GetName()}); status.Code(err) != codes.NotFound {
			t.Fatalf("expected record to be deleted, got: %+v, %v", r, err)
		}
	})

	t.Run("already deleted", func(t *testing.T) {
		// Check if a deleted record can be deleted again
		if _, err := srv.DeleteRecord(ctx, &rpb.DeleteRecordRequest{Name: r.GetName()}); status.Code(err) != codes.NotFound {
			t.Fatalf("expected NOT_FOUND, got: %v", err)
		}
	})
}

// TestListRecords_multiresult tests listing records across multiple parents.
func TestListRecords_multiresult(t *testing.T) {
	// Create a temporary database
	srv, err := New(test.NewDB(t))
	if err != nil {
		t.Fatalf("failed to setup db: %v", err)
	}
	ctx := context.Background()

	records := make([]*rpb.Record, 0, 8)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			res, err := srv.CreateResult(ctx, &rpb.CreateResultRequest{
				Parent: strconv.Itoa(i),
				Result: &rpb.Result{
					Name: fmt.Sprintf("%d/results/%d", i, j),
				},
			})
			if err != nil {
				t.Fatalf("CreateResult(): %v", err)
			}
			for k := 0; k < 2; k++ {
				r, err := srv.CreateRecord(ctx, &rpb.CreateRecordRequest{
					Parent: res.GetName(),
					Record: &rpb.Record{
						Name: record.FormatName(res.GetName(), strconv.Itoa(k)),
					},
				})
				if err != nil {
					t.Fatalf("CreateRecord(): %v", err)
				}
				records = append(records, r)
			}
		}
	}

	got, err := srv.ListRecords(ctx, &rpb.ListRecordsRequest{
		Parent: "0/results/-",
	})
	if err != nil {
		t.Fatalf("ListRecords(): %v", err)
	}
	want := &rpb.ListRecordsResponse{
		Records: records[:4],
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Error(diff)
	}
}
