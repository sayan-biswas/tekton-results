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

package record

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	cw "github.com/jonboulle/clockwork"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/results/pkg/api/server/db"
	"github.com/tektoncd/results/pkg/api/server/v1alpha2/log"
	"github.com/tektoncd/results/pkg/apis/v1alpha2"
	"github.com/tektoncd/results/pkg/conf"
	"github.com/tektoncd/results/pkg/internal/jsonutil"
	ppb "github.com/tektoncd/results/proto/pipeline/v1beta1/pipeline_go_proto"
	pb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
	"time"
)

var clock cw.Clock = cw.NewFakeClock()

func TestParseName(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		// if want is nil, assume error
		want []string
	}{
		{
			name: "simple",
			in:   "clusters/a/namespaces/b/results/c/records/d",
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "resource name reuse",
			in:   "clusters/clusters/namespaces/namespaces/results/results/records/records",
			want: []string{"clusters", "namespaces", "results", "records"},
		},
		{
			name: "missing name",
			in:   "clusters/a/namespaces/b/results/c/records/",
		},
		{
			name: "missing name, no slash",
			in:   "clusters/a/namespaces/b/results/c/records",
		},
		{
			name: "missing parent",
			in:   "/records/b",
		},
		{
			name: "missing parent, no slash",
			in:   "records/b",
		},
		{
			name: "wrong resource",
			in:   "clusters/a/namespaces/b/wrong/c/records/c",
		},
		{
			name: "result resource",
			in:   "clusters/a/namespaces/b/results/c",
		},
		{
			name: "invalid parent",
			in:   "a/b/results/c",
		},
		{
			name: "invalid name",
			in:   "clusters/a/namespaces/b/results/c/records/d/e",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cluster, namespace, result, name, err := ParseName(tc.in)
			if err != nil {
				if tc.want == nil {
					// error was expected, continue
					return
				}
				t.Fatal(err)
			}
			if tc.want == nil {
				t.Fatalf("expected error, got: [%s, %s, %s, %s]", cluster, namespace, result, name)
			}
			if cluster != tc.want[0] || namespace != tc.want[1] || result != tc.want[2] || name != tc.want[3] {
				t.Errorf("want: %v, got: [%s, %s, %s, %s]", tc.want, cluster, namespace, result, name)
			}
		})
	}
}

func TestParseParent(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		// if want is nil, assume error
		want []string
	}{
		{
			name: "simple",
			in:   "clusters/a/namespaces/b/results/c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "resource name reuse",
			in:   "clusters/clusters/namespaces/namespaces/results/results",
			want: []string{"clusters", "namespaces", "results"},
		},
		{
			name: "missing name",
			in:   "clusters/a/namespaces/b/results/",
		},
		{
			name: "missing name, no slash",
			in:   "clusters/a/namespaces/b/results",
		},
		{
			name: "wrong resource",
			in:   "clusters/a/namespaces/b/wrong/c",
		},
		{
			name: "no result resource",
			in:   "clusters/a/namespaces/b",
		},
		{
			name: "invalid parent",
			in:   "a/b/results/c",
		},
		{
			name: "invalid name",
			in:   "clusters/a/namespaces/b/results/c/d",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cluster, namespace, result, err := ParseParent(tc.in)
			if err != nil {
				if tc.want == nil {
					// error was expected, continue
					return
				}
				t.Fatal(err)
			}
			if tc.want == nil {
				t.Fatalf("expected error, got: [%s, %s, %s]", cluster, namespace, result)
			}
			if cluster != tc.want[0] || namespace != tc.want[1] || result != tc.want[2] {
				t.Errorf("want: %v, got: [%s, %s, %s]", tc.want, cluster, namespace, result)
			}
		})
	}
}

func TestParseParentDB(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		// if want is nil, assume error
		want []string
	}{
		{
			name: "simple",
			in:   "a/b",
			want: []string{"a", "b"},
		},
		{
			name: "missing cluster",
			in:   "/a",
		},
		{
			name: "missing namespace",
			in:   "a/",
		},
		{
			name: "missing namespace, no slash",
			in:   "a",
		},
		{
			name: "less parameters",
			in:   "abc",
		},
		{
			name: "more parameters",
			in:   "a/b/c",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cluster, namespace, err := ParseParentDB(tc.in)
			if err != nil {
				if tc.want == nil {
					// error was expected, continue
					return
				}
				t.Fatal(err)
			}
			if tc.want == nil {
				t.Fatalf("expected error, got: [%s, %s]", cluster, namespace)
			}
			if cluster != tc.want[0] || namespace != tc.want[1] {
				t.Errorf("want: %v, got: [%s, %s]", tc.want, cluster, namespace)
			}
		})
	}
}

func TestToStorage(t *testing.T) {
	data := &ppb.TaskRun{Metadata: &ppb.ObjectMeta{Name: "tacocat"}}

	for _, tc := range []struct {
		name string
		in   *pb.Record
		want *db.Record
	}{
		{
			name: "full",
			in: &pb.Record{
				Name: "clusters/a/namespaces/b/results/c/records/d",
				Id:   "1",
				Data: &pb.Any{
					Value: jsonutil.AnyBytes(t, data),
				},
				CreatedTime: timestamppb.New(clock.Now()),
				UpdatedTime: timestamppb.New(clock.Now()),
				Etag:        "tacocat",
			},
			want: &db.Record{
				Parent:      "a/b",
				ResultID:    "1",
				ResultName:  "c",
				Name:        "d",
				ID:          "1",
				Data:        jsonutil.AnyBytes(t, data),
				CreatedTime: clock.Now(),
				UpdatedTime: clock.Now(),
				Etag:        "tacocat",
			},
		},
		{
			name: "missing data",
			in: &pb.Record{
				Name:       "clusters/a/namespaces/b/results/c/records/d",
				Id:         "1",
				CreateTime: timestamppb.New(clock.Now()),
			},
			want: &db.Record{
				Parent:      "a/b",
				ResultID:    "1",
				ResultName:  "c",
				Name:        "d",
				ID:          "1",
				CreatedTime: clock.Now(),
			},
		},
		{
			name: "deprecated fields", // If deprecated fields do not match their non-deprecated counterparts, prefer non-deprecated.
			in: &pb.Record{
				Name: "clusters/a/namespaces/b/results/c/records/d",
				Uid:  "1",
				Id:   "2",
				Data: &pb.Any{
					Value: jsonutil.AnyBytes(t, data),
				},
				CreatedTime: timestamppb.New(clock.Now().Add(1 * time.Minute)),
				CreateTime:  timestamppb.New(clock.Now()),
				UpdatedTime: timestamppb.New(clock.Now().Add(1 * time.Minute)),
				UpdateTime:  timestamppb.New(clock.Now()),
				Etag:        "tacocat",
			},
			want: &db.Record{
				Parent:      "a/b",
				ResultID:    "1",
				ResultName:  "c",
				Name:        "d",
				ID:          "1",
				Data:        jsonutil.AnyBytes(t, data),
				CreatedTime: clock.Now(),
				UpdatedTime: clock.Now(),
				Etag:        "tacocat",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToStorage("a", "b", "c", "1", "d", tc.in, nil)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("-want,+got: %s", diff)
			}
		})
	}

	// errors
	for _, tc := range []struct {
		name string
		in   *pb.Record
		want codes.Code
	}{
		{
			name: "invalid type",
			in: &pb.Record{
				Name: "clusters/a/namespaces/b/results/c",
				Id:   "1",
				Data: &pb.Any{
					Type: strings.Repeat("a", typeSize+1),
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "invalid data",
			in: &pb.Record{
				Name: "clusters/a/namespaces/b/results/c",
				Id:   "1",
				Data: &pb.Any{},
			},
			want: codes.InvalidArgument,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToStorage("a", "b", "c", "1", "d", tc.in, nil)
			if status.Code(err) != tc.want {
				t.Fatalf("expected %v, got (%v, %v)", tc.want, got, err)
			}
		})
	}
}

func TestToAPI(t *testing.T) {
	data := &v1beta1.TaskRun{ObjectMeta: v1.ObjectMeta{Name: "tacocat"}}
	for _, tc := range []struct {
		name string
		in   *db.Record
		want *pb.Record
	}{
		{
			name: "full",
			in: &db.Record{
				Parent:      "a/b",
				ResultID:    "1",
				ResultName:  "c",
				Name:        "d",
				ID:          "1",
				Data:        jsonutil.AnyBytes(t, data),
				CreatedTime: clock.Now(),
				Etag:        "etag",
			},
			want: &pb.Record{
				Name: "clusters/a/namespaces/b/results/c/records/d",
				Id:   "1",
				Uid:  "1",
				Data: &pb.Any{
					Value: jsonutil.AnyBytes(t, data),
				},
				CreatedTime: timestamppb.New(clock.Now()),
				CreateTime:  timestamppb.New(clock.Now()),
				Etag:        "etag",
			},
		},
		{
			name: "partial",
			in: &db.Record{
				Parent:     "a/b",
				ResultID:   "1",
				ResultName: "c",
				Name:       "d",
				ID:         "1",
			},
			want: &pb.Record{
				Name: "clusters/a/namespaces/b/results/c/records/d",
				Id:   "1",
				Uid:  "1",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToAPI(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("-want,+got: %s", diff)
			}
		})
	}
}

type mockLogStreamer struct {
	streamerType string
}

func (s *mockLogStreamer) WriteTo(io.Writer) (int64, error) {
	return 0, fmt.Errorf("not implemented!")
}

func (s *mockLogStreamer) ReadFrom(io.Reader) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (s *mockLogStreamer) Type() string {
	return s.streamerType
}

func TestToLogStreamer(t *testing.T) {
	cases := []struct {
		name      string
		in        *db.Record
		want      log.LogStreamer
		expectErr bool
	}{
		{
			name: "TaskRunLog Filesystem Type",
			in: &db.Record{
				Parent:     "app",
				ResultID:   "1",
				ResultName: "push-main",
				Name:       "taskrun-compile-log",
				ID:         "a",
				Type:       v1alpha2.TaskRunLogRecordType,
				Data: jsonutil.AnyBytes(t, &v1alpha2.TaskRunLog{
					Spec: v1alpha2.TaskRunLogSpec{
						Type: v1alpha2.FileLogType,
						Ref: v1alpha2.TaskRunRef{
							Namespace: "app",
							Name:      "taskrun-compile",
						},
					},
				}),
			},
			want: &mockLogStreamer{
				streamerType: string(v1alpha2.FileLogType),
			},
		},
		{
			name: "TaskRun Record",
			in: &db.Record{
				Parent:     "app",
				ResultID:   "1",
				ResultName: "push-main",
				Name:       "taskrun-compile",
				ID:         "a",
				Type:       "pipeline.tekton.dev/TaskRun",
			},
			expectErr: true,
		},
		{
			name: "PipelineRun Record",
			in: &db.Record{
				Parent:     "app",
				ResultID:   "1",
				ResultName: "push-main",
				Name:       "taskrun-compile",
				ID:         "a",
				Type:       "pipeline.tekton.dev/PipelineRun",
			},
			expectErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			streamer, _, err := ToLogStreamer(tc.in, 1024, "", &conf.ConfigFile{}, context.TODO())
			if err != nil {
				if !tc.expectErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if streamer.Type() != tc.want.Type() {
				t.Errorf("expected log streamer %s, got %s", tc.want.Type(), streamer.Type())
			}
		})
	}
}

func TestFormatName(t *testing.T) {
	got := FormatName("a", "b")
	want := "a/records/b"
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestFormatParent(t *testing.T) {
	got := FormatParent("a", "b", "c")
	want := "clusters/a/namespaces/b/results/c"
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestFormatParentDB(t *testing.T) {
	got := FormatParentDB("a", "b")
	want := "a/b"
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestValidateType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := ValidateType("foo"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		if err := ValidateType(strings.Repeat("a", typeSize+1)); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
