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
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/tektoncd/results/pkg/internal/protoutil"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2/record"
	"github.com/tektoncd/results/pkg/server/auth"
	celenv "github.com/tektoncd/results/pkg/server/cel"
	"github.com/tektoncd/results/pkg/server/db/errors"
	"github.com/tektoncd/results/pkg/server/db/models"
	"github.com/tektoncd/results/pkg/server/db/ordering"
	"github.com/tektoncd/results/pkg/server/db/pagination"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// GetRecord returns a single Record.
func (s *Server) GetRecord(ctx context.Context, req *rpb.GetRecordRequest) (*rpb.Record, error) {
	//Parse input request
	cluster, namespace, resultName, name, err := record.ParseName(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceRecords, auth.PermissionGet); err != nil {
		return nil, err
	}

	// Format parent name
	parent := record.FormatParentDB(cluster, namespace)

	// Retrieve record from storage
	r, err := record.GetRecord(ctx, s.db.WithContext(ctx), parent, resultName, name)
	if err != nil {
		return nil, err
	}

	return record.ToAPI(r)
}

// ListRecords returns collection of records in a particular workspace, namespace and results.
func (s *Server) ListRecords(ctx context.Context, req *rpb.ListRecordsRequest) (*rpb.ListRecordsResponse, error) {
	//Parse input request
	cluster, namespace, resultName, err := record.ParseParent(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceRecords, auth.PermissionList); err != nil {
		return nil, err
	}

	// Format parent name
	parent := record.FormatParentDB(cluster, namespace)

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

	env, err := record.RecordCEL()
	if err != nil {
		return nil, err
	}
	prg, err := celenv.ParseFilter(env, req.GetFilter())
	if err != nil {
		return nil, err
	}
	// Fetch n+1 items to get the next token.
	r, err := record.GetFilteredPaginatedSortedRecords(ctx, s.db, parent, resultName, start, userPageSize+1, prg, sortOrder)
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

	return &rpb.ListRecordsResponse{
		Records:       r,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) CreateRecord(ctx context.Context, req *rpb.CreateRecordRequest) (*rpb.Record, error) {
	r := req.GetRecord()

	//Parse input request
	cluster, namespace, name, err := record.ParseParent(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceRecords, auth.PermissionCreate); err != nil {
		return nil, err
	}

	// Format parent name
	parent := record.FormatParentDB(cluster, namespace)

	// Look up the result ID from the name. This does not have to happen
	// transactionally with the insert since name<->ID mappings are immutable,
	// and if the the parent result is deleted mid-request, the insert should
	// fail due to foreign key constraints.
	resultID, err := s.getResultID(ctx, s.db, parent, name)
	if err != nil {
		return nil, err
	}

	// Populate Result with server provided fields.
	protoutil.ClearOutputOnly(r)
	r.Id = uid()
	ts := timestamppb.New(clock.Now())
	r.CreateTime = ts
	r.UpdateTime = ts

	// Insert record in storage
	store, err := record.ToStorage(r)
	if err != nil {
		return nil, err
	}
	store.ResultID = resultID
	if err := record.UpdateEtag(store); err != nil {
		return nil, err
	}
	q := s.db.WithContext(ctx).
		Model(store).
		Create(store).Error
	if err := errors.Wrap(q); err != nil {
		return nil, err
	}

	return record.ToAPI(store)
}

// UpdateRecord updates a record in the database.
func (s *Server) UpdateRecord(ctx context.Context, req *rpb.UpdateRecordRequest) (*rpb.Record, error) {
	rec := req.GetRecord()

	//Parse input request
	cluster, namespace, resultName, name, err := record.ParseName(rec.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceRecords, auth.PermissionUpdate); err != nil {
		return nil, err
	}

	// Format parent name
	parent := record.FormatParentDB(cluster, namespace)

	protoutil.ClearOutputOnly(rec)

	var out *rpb.Record
	err = s.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		r, err := record.GetRecord(ctx, db, parent, resultName, name)
		if err != nil {
			return err
		}

		// If the user provided the Etag field, then make sure the value of this field matches what saved in the database.
		// See https://google.aip.dev/154 for more information.
		if req.GetEtag() != "" && req.GetEtag() != r.Etag {
			return status.Error(codes.FailedPrecondition, "the etag mismatches")
		}

		// Merge existing data with user request.
		pb, err := record.ToAPI(r)
		if err != nil {
			return err
		}
		// TODO: field mask support.
		proto.Merge(pb, rec)

		updateTime := timestamppb.New(clock.Now())
		pb.UpdateTime = updateTime

		// Convert back to storage and store.
		s, err := record.ToStorage(pb)
		if err != nil {
			return err
		}
		s.ResultID = r.ResultID
		if err := record.UpdateEtag(s); err != nil {
			return err
		}
		if err := errors.Wrap(db.Save(s).Error); err != nil {
			return err
		}

		pb.Etag = s.Etag
		out = pb
		return nil
	})
	return out, err
}

// DeleteRecord deletes a given record.
func (s *Server) DeleteRecord(ctx context.Context, req *rpb.DeleteRecordRequest) (*empty.Empty, error) {
	//Parse input request
	cluster, namespace, resultName, name, err := record.ParseName(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check access
	if err := s.auth.Check(ctx, cluster, namespace, auth.ResourceRecords, auth.PermissionDelete); err != nil {
		return nil, err
	}

	// Format parent name
	parent := record.FormatParentDB(cluster, namespace)

	// First get the current record. This ensures that we return NOT_FOUND if
	// the entry is already deleted.
	// This does not need to be done in the same transaction as the delete,
	// since the identifiers are immutable.
	r, err := record.GetRecord(ctx, s.db, parent, resultName, name)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, errors.Wrap(s.db.WithContext(ctx).Delete(&models.Record{}, r).Error)
}
