// Copyright 2021 The Tekton Authors
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

package auth_test

import (
	"context"
	"fmt"
	"testing"

	testclient "github.com/tektoncd/results/pkg/internal/test"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2"
	"github.com/tektoncd/results/pkg/server/auth"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	test "k8s.io/client-go/testing"
)

func TestRBAC(t *testing.T) {
	// Map of users -> tokens. The 'authorized' user has full permissions.
	users := map[string]string{
		"authorized":   "a",
		"unauthorized": "b",
	}
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor("create", "subjectaccessreviews", func(action test.Action) (handled bool, ret runtime.Object, err error) {
		sar := action.(test.CreateActionImpl).Object.(*authzv1.SubjectAccessReview)
		if sar.Spec.User == "authorized" {
			sar.Status.Allowed = true
		} else {
			sar.Status.Denied = true
		}
		return true, sar, nil
	})
	client := testclient.NewResultsClient(t, v1alpha2.WithAuth(auth.NewRBAC()))

	ctx := context.Background()
	result := "foo/results/bar"
	record := "foo/results/bar/records/baz"
	for _, tc := range []struct {
		user  string
		token string
		want  codes.Code
	}{
		{
			user:  "authorized",
			token: users["authorized"],
			want:  codes.OK,
		},
		{
			user:  "unauthorized",
			token: users["unauthorized"],
			want:  codes.Unauthenticated,
		},
		{
			user:  "unauthenticated",
			token: "",
			want:  codes.Unauthenticated,
		},
	} {
		t.Run(tc.user, func(t *testing.T) {
			// Simulates a oauth.TokenSource. We avoid using the real
			// oauth.TokenSource here since it requires a higher SecurityLevel
			// + TLS.
			ctx := metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", tc.token))
			if _, err := client.CreateResult(ctx, &rpb.CreateResultRequest{
				Parent: "foo",
				Result: &rpb.Result{
					Name: "foo/results/bar",
				},
			}); status.Code(err) != tc.want {
				t.Fatalf("CreateResult: %v, want %v", err, tc.want)
			}
			if _, err := client.GetResult(ctx, &rpb.GetResultRequest{Name: result}); status.Code(err) != tc.want {
				t.Fatalf("GetResult: %v, want %v", err, tc.want)
			}
			if _, err := client.ListResults(ctx, &rpb.ListResultsRequest{Parent: "foo"}); status.Code(err) != tc.want {
				t.Fatalf("ListResult: %v, want %v", err, tc.want)
			}
			if _, err := client.UpdateResult(ctx, &rpb.UpdateResultRequest{Result: &rpb.Result{Name: result}}); status.Code(err) != tc.want {
				t.Fatalf("UpdateResult: %v, want %v", err, tc.want)
			}

			if _, err := client.CreateRecord(ctx, &rpb.CreateRecordRequest{
				Parent: result,
				Record: &rpb.Record{
					Name: record,
				},
			}); status.Code(err) != tc.want {
				t.Fatalf("CreateRecord: %v, want %v", err, tc.want)
			}
			if _, err := client.GetRecord(ctx, &rpb.GetRecordRequest{Name: record}); status.Code(err) != tc.want {
				t.Fatalf("GetRecord: %v, want %v", err, tc.want)
			}
			if _, err := client.ListRecords(ctx, &rpb.ListRecordsRequest{Parent: result}); status.Code(err) != tc.want {
				t.Fatalf("ListRecord: %v, want %v", err, tc.want)
			}
			if _, err := client.UpdateRecord(ctx, &rpb.UpdateRecordRequest{Record: &rpb.Record{Name: record}}); status.Code(err) != tc.want {
				t.Fatalf("UpdateRecord: %v, want %v", err, tc.want)
			}

			if _, err := client.DeleteRecord(ctx, &rpb.DeleteRecordRequest{Name: record}); status.Code(err) != tc.want {
				t.Fatalf("DeleteRecord: %v, want %v", err, tc.want)
			}
			if _, err := client.DeleteResult(ctx, &rpb.DeleteResultRequest{Name: result}); status.Code(err) != tc.want {
				t.Fatalf("DeleteResult: %v, want %v", err, tc.want)
			}
		})
	}
}
