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

package result

import (
	"github.com/google/go-cmp/cmp"
	"knative.dev/pkg/ptr"
	"strings"
	"testing"
	"time"

	cw "github.com/jonboulle/clockwork"
	"github.com/tektoncd/results/pkg/api/server/cel"
	"github.com/tektoncd/results/pkg/api/server/db"
	pb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			name: "missing parent",
			in:   "/results/a",
		},
		{
			name: "missing parent, no slash",
			in:   "results/a",
		},
		{
			name: "wrong resource",
			in:   "clusters/a/namespaces/b/record/c",
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
			cluster, namespace, name, err := ParseName(tc.in)
			if err != nil {
				if tc.want == nil {
					// error was expected, continue
					return
				}
				t.Fatal(err)
			}
			if tc.want == nil {
				t.Fatalf("expected error, got: [%s, %s, %s]", cluster, namespace, name)
			}

			if cluster != tc.want[0] || namespace != tc.want[1] || name != tc.want[2] {
				t.Errorf("want: %v, got: [%s, %s, %s]", tc.want, cluster, namespace, name)
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
			in:   "clusters/a/namespaces/b",
			want: []string{"a", "b"},
		},
		{
			name: "resource name reuse",
			in:   "clusters/clusters/namespaces/namespaces",
			want: []string{"clusters", "namespaces"},
		},
		{
			name: "missing namespace",
			in:   "clusters/a/namespaces/",
		},
		{
			name: "missing name, no slash",
			in:   "clusters/a/namespaces",
		},
		{
			name: "invalid namespace",
			in:   "clusters/a/namespaces/b/c",
		},
		{
			name: "missing cluster",
			in:   "clusters//namespaces/b",
		},
		{
			name: "missing cluster, no slash",
			in:   "clusters/namespaces/b",
		},
		{
			name: "invalid cluster",
			in:   "clusters/a/b/namespaces/c",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cluster, namespace, err := ParseParent(tc.in)
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
	for _, tc := range []struct {
		name string
		in   *pb.Result
		want *db.Result
	}{
		{
			name: "all",
			in: &pb.Result{
				Name:        "clusters/a/namespaces/b/results/c",
				Id:          "1",
				CreatedTime: timestamppb.New(clock.Now()),
				UpdatedTime: timestamppb.New(clock.Now()),
				Annotations: map[string]string{"a": "b"},
				Etag:        "tacocat",
				Summary: &pb.RecordSummary{
					Record:      "clusters/a/namespaces/b/results/c/records/d",
					Type:        "type",
					StartTime:   timestamppb.New(clock.Now()),
					EndTime:     timestamppb.New(clock.Now()),
					Status:      pb.RecordSummary_SUCCESS,
					Annotations: map[string]string{"c": "d"},
				},
			},
			want: &db.Result{
				Parent:      "a/b",
				Name:        "c",
				ID:          "1",
				Annotations: map[string]string{"a": "b"},
				CreatedTime: clock.Now(),
				UpdatedTime: clock.Now(),
				Etag:        "tacocat",
				Summary: db.RecordSummary{
					Record:      "clusters/a/namespaces/b/results/c/records/d",
					Type:        "type",
					StartTime:   ptr.Time(clock.Now()),
					EndTime:     ptr.Time(clock.Now()),
					Status:      1,
					Annotations: map[string]string{"c": "d"},
				},
			},
		},
		{
			name: "deprecated fields",
			in: &pb.Result{
				Name:        "clusters/a/namespaces/b/results/c",
				Uid:         "1",
				Id:          "2",
				CreatedTime: timestamppb.New(clock.Now().Add(time.Minute)),
				CreateTime:  timestamppb.New(clock.Now()),
				UpdatedTime: timestamppb.New(clock.Now().Add(time.Minute)),
				UpdateTime:  timestamppb.New(clock.Now()),
				Annotations: map[string]string{"a": "b"},
				Etag:        "tacocat",
			},
			want: &db.Result{
				Parent:      "a/b",
				Name:        "c",
				ID:          "1",
				Annotations: map[string]string{"a": "b"},
				CreatedTime: clock.Now(),
				UpdatedTime: clock.Now(),
				Etag:        "tacocat",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToStorage(tc.in)
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
		in   *pb.Result
		want codes.Code
	}{
		{
			name: "invalid summary record name",
			in: &pb.Result{
				Name: "clusters/a/namespaces/b/results/c",
				Id:   "1",
				Summary: &pb.RecordSummary{
					Record: "d",
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "invalid summary type",
			in: &pb.Result{
				Name: "clusters/a/namespaces/b/results/c",
				Id:   "1",
				Summary: &pb.RecordSummary{
					Record: "clusters/a/namespaces/b/results/c/records/d",
					Type:   strings.Repeat("a", 1024),
				},
			},
			want: codes.InvalidArgument,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToStorage(tc.in)
			if status.Code(err) != tc.want {
				t.Fatalf("expected %v, got (%v, %v)", tc.want, got, err)
			}
		})
	}
}

func TestToAPI(t *testing.T) {
	ann := map[string]string{"a": "b"}
	got, err := ToAPI(&db.Result{
		Parent:      "a/b",
		Name:        "c",
		ID:          "1",
		CreatedTime: clock.Now(),
		UpdatedTime: clock.Now(),
		Annotations: ann,
		Etag:        "etag",
	})
	want := &pb.Result{
		Name:        "clusters/a/namespaces/b/results/c",
		Id:          "1",
		Uid:         "1",
		CreatedTime: timestamppb.New(clock.Now()),
		CreateTime:  timestamppb.New(clock.Now()),
		UpdatedTime: timestamppb.New(clock.Now()),
		UpdateTime:  timestamppb.New(clock.Now()),
		Annotations: ann,
		Etag:        "etag",
	}
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("-want,+got: %s", diff)
	}
}

func TestMatch(t *testing.T) {
	env, err := cel.NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	r := &pb.Result{
		Name:        "foo",
		Id:          "bar",
		CreateTime:  timestamppb.Now(),
		Annotations: map[string]string{"a": "b"},
		Etag:        "tacocat",
	}
	for _, tc := range []struct {
		name   string
		result *pb.Result
		filter string
		match  bool
		status codes.Code
	}{
		{
			name:   "no filter",
			filter: "",
			result: r,
			match:  true,
		},
		{
			name:   "matching condition",
			filter: `result.id != ""`,
			result: r,
			match:  true,
		},
		{
			name:   "non-matching condition",
			filter: `result.id == ""`,
			result: r,
			match:  false,
		},
		{
			name:   "nil result",
			result: nil,
			filter: "result.id",
			match:  false,
		},
		{
			name:   "non-bool output",
			result: r,
			filter: "result",
			status: codes.InvalidArgument,
		},
		{
			name:   "wrong resource type",
			result: r,
			filter: "record",
			status: codes.InvalidArgument,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := cel.ParseFilter(env, tc.filter)
			if err != nil {
				t.Fatalf("ParseFilter: %v", err)
			}
			got, err := Match(tc.result, p)
			if status.Code(err) != tc.status {
				t.Fatalf("Match: %v", err)
			}
			if got != tc.match {
				t.Errorf("want: %t, got: %t", tc.match, got)
			}
		})
	}
}

func TestFormatName(t *testing.T) {
	got := FormatName("a", "b")
	want := "a/results/b"
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestFormatParent(t *testing.T) {
	got := FormatParent("a", "b")
	want := "clusters/a/namespaces/b"
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
