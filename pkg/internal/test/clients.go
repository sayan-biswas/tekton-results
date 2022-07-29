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

package test

import (
	"fmt"
	"net"
	"testing"

	"github.com/tektoncd/results/pkg/server/api/v1alpha2"
	"github.com/tektoncd/results/pkg/server/db/test"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/grpc"
)

const (
	port = ":0"
)

func NewResultsClient(t *testing.T, opts ...v1alpha2.Option) rpb.ResultsClient {
	t.Helper()
	srv, err := v1alpha2.New(test.NewDB(t), opts...)
	if err != nil {
		t.Fatalf("Failed to create fake server: %v", err)
	}
	s := grpc.NewServer()
	rpb.RegisterResultsServer(s, srv) // local test server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			fmt.Printf("error starting result server: %v\n", err)
		}
	}()
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	t.Cleanup(func() {
		s.Stop()
		lis.Close()
		conn.Close()
	})
	return rpb.NewResultsClient(conn)
}
