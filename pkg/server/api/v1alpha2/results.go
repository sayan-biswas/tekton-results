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

package v1alpha2

import (
	"context"
	celenv "github.com/tektoncd/results/pkg/server/cel"
	"github.com/tektoncd/results/pkg/server/db/ordering"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"gorm.io/gorm"

	"github.com/tektoncd/results/pkg/internal/protoutil"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2/result"
	"github.com/tektoncd/results/pkg/server/auth"
	"github.com/tektoncd/results/pkg/server/db/errors"
	"github.com/tektoncd/results/pkg/server/db/models"
	"github.com/tektoncd/results/pkg/server/db/pagination"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetResult returns a single Result.
func (s *Server) GetResult(ctx context.Context, req *rpb.GetResultRequest) (*rpb.Result, error) {
	//Parse input request
	cluster, namespace, name, err := result.ParseName(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceResults, auth.PermissionGet); err != nil {
		return nil, err
	}

	// Format parent name
	parent := result.FormatParentDB(cluster, namespace)

	// Query database
	store, err := result.GetResult(ctx, s.db, parent, name)
	if err != nil {
		return nil, err
	}

	return result.ToAPI(store), nil
}

// ListResults returns collection of results in a particular workspace and namespace.
func (s *Server) ListResults(ctx context.Context, req *rpb.ListResultsRequest) (*rpb.ListResultsResponse, error) {
	//Parse input request
	cluster, namespace, err := result.ParseParent(req.Parent)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceResults, auth.PermissionList); err != nil {
		return nil, err
	}

	userPageSize, err := pagination.PageSize(int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}

	start, err := pagination.PageStart(req.GetPageToken(), req.GetFilter())
	if err != nil {
		return nil, err
	}

	sortOrder, err := ordering.OrderBy(req.GetOrderBy())
	if err != nil {
		return nil, err
	}

	prg, err := celenv.ParseFilter(s.env, req.GetFilter())
	if err != nil {
		return nil, err
	}

	// Format parent name
	parent := result.FormatParentDB(cluster, namespace)

	// Fetch n+1 items to get the next token.
	r, err := result.GetFilteredPaginatedSortedResults(ctx, s.db, parent, start, userPageSize+1, prg, sortOrder)
	if err != nil {
		return nil, err
	}

	// If we returned the full n+1 items, use the last element as the next page
	// token.
	var nextToken string
	if len(r) > userPageSize {
		next := r[len(r)-1]
		var err error
		nextToken, err = pagination.EncodeToken(next.GetId(), req.GetFilter())
		if err != nil {
			return nil, err
		}
		r = r[:len(r)-1]
	}

	return &rpb.ListResultsResponse{
		Results:       r,
		NextPageToken: nextToken,
	}, nil
}

// CreateResult creates a new result in the database.
func (s *Server) CreateResult(ctx context.Context, req *rpb.CreateResultRequest) (*rpb.Result, error) {
	r := req.GetResult()

	//Parse input request
	cluster, namespace, err := result.ParseParent(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceResults, auth.PermissionCreate); err != nil {
		return nil, err
	}

	// Populate Result with server provided fields.
	protoutil.ClearOutputOnly(r)
	r.Id = uid()
	ts := timestamppb.New(clock.Now())
	r.CreateTime = ts
	r.UpdateTime = ts

	// Insert in database
	store, err := result.ToStorage(r)
	if err != nil {
		return nil, err
	}

	if err := result.UpdateEtag(store); err != nil {
		return nil, err
	}

	if err := errors.Wrap(s.db.WithContext(ctx).Create(store).Error); err != nil {
		return nil, err
	}

	return result.ToAPI(store), nil
}

// UpdateResult updates a Result in the database.
func (s *Server) UpdateResult(ctx context.Context, req *rpb.UpdateResultRequest) (*rpb.Result, error) {
	res := req.GetResult()

	//Parse input request
	cluster, namespace, name, err := result.ParseName(res.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceResults, auth.PermissionUpdate); err != nil {
		return nil, err
	}

	// Format parent name
	parent := result.FormatParentDB(cluster, namespace)

	// Update database
	var r *rpb.Result
	err = s.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		prev, err := result.GetResult(ctx, db, parent, name)
		if err != nil {
			return status.Errorf(codes.NotFound, "failed to find a result: %v", err)
		}

		// If the user provided the Etag field, then make sure the value of this field matches what saved in the database.
		// See https://google.aip.dev/154 for more information.
		if req.GetEtag() != "" && req.GetEtag() != prev.Etag {
			return status.Error(codes.FailedPrecondition, "the etag mismatches")
		}

		newpb := result.ToAPI(prev)
		reqpb := req.GetResult()
		protoutil.ClearOutputOnly(reqpb)
		// Merge requested Result with previous Result to apply updates,
		// making sure to filter out any OUTPUT_ONLY fields, and only
		// updatable fields.
		// We can't use proto.Merge, since empty fields in the req should take
		// precedence, so set each updatable field here.
		newpb.Annotations = reqpb.GetAnnotations()
		newpb.Summary = reqpb.GetSummary()
		toDB, err := result.ToStorage(newpb)
		if err != nil {
			return err
		}

		// Set server-side provided fields
		toDB.UpdatedTime = clock.Now()
		if err := result.UpdateEtag(toDB); err != nil {
			return err
		}

		// Write result back to database.
		if err = errors.Wrap(db.Save(toDB).Error); err != nil {
			log.Printf("failed to save result into database: %v", err)
			return err
		}
		r = result.ToAPI(toDB)
		return nil
	})
	return r, err
}

// DeleteResult deletes a given result.
func (s *Server) DeleteResult(ctx context.Context, req *rpb.DeleteResultRequest) (*empty.Empty, error) {
	//Parse input request
	cluster, namespace, name, err := result.ParseName(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceResults, auth.PermissionDelete); err != nil {
		return nil, err
	}

	// Format parent name
	parent := result.FormatParentDB(cluster, namespace)

	// Check whether result exists
	r, err := result.GetResult(ctx, s.db, parent, name)
	if err != nil {
		return &empty.Empty{}, err
	}

	// Delete the result.
	db := s.db.WithContext(ctx).Delete(&models.Result{}, r)
	return &empty.Empty{}, errors.Wrap(db.Error)
}
