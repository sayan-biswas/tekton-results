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

package protoutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestClearOutputOnly(t *testing.T) {
	m := &rpb.Result{
		Name:        "a",
		Id:          "b",
		CreatedTime: timestamppb.Now(),
		UpdatedTime: timestamppb.Now(),
		Annotations: map[string]string{"c": "d"},
		Etag:        "f",
	}
	want := &rpb.Result{
		Name:        m.Name,
		Annotations: m.Annotations,
	}

	ClearOutputOnly(m)

	if diff := cmp.Diff(want, m, protocmp.Transform()); diff != "" {
		t.Errorf("-want, +got: %s", diff)
	}
}
