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

// Package result provides utilities for manipulating and validating Results.
package result

import (
	"context"
	"fmt"
	"github.com/tektoncd/results/pkg/server/db/errors"
	"github.com/tektoncd/results/pkg/server/db/pagination"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2/record"
	celenv "github.com/tektoncd/results/pkg/server/cel"
	"github.com/tektoncd/results/pkg/server/db/models"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"knative.dev/pkg/ptr"
)

// resultID is a utility struct to extract partial Result data representing
// Result name <-> ID mappings.
type resultID struct {
	Name string
	ID   string
}

var (
	// NameRegex matches valid name specs for a Result.
	NameRegex   = regexp.MustCompile("^clusters/([a-z0-9:_-]{1,63})/namespaces/([a-z0-9_-]{1,63})/results/([a-z0-9_-]{1,63})$")
	ParentRegex = regexp.MustCompile("^clusters/([a-z0-9:_-]{1,63})/namespaces/([a-z0-9_-]{1,63})$")
)

// ParseParent splits a top-level parent into its individual (cluster, namespace)
func ParseParent(raw string) (cluster, namespace string, err error) {
	s := ParentRegex.FindStringSubmatch(raw)
	if len(s) != 3 {
		return "", "", status.Errorf(codes.InvalidArgument, "parent must match %s", ParentRegex.String())
	}
	return s[1], s[2], nil
}

// ParseName splits a full Result name into its individual (parent, name)
// components.
func ParseName(raw string) (cluster, namespace, name string, err error) {
	s := NameRegex.FindStringSubmatch(raw)
	if len(s) != 4 {
		return "", "", "", status.Errorf(codes.InvalidArgument, "name must match %s", NameRegex.String())
	}
	return s[1], s[2], s[3], nil
}

// ParseParentDB splits database field parent into its individual (cluster, namespace)
func ParseParentDB(raw string) (cluster, namespace string) {
	s := strings.Split(raw, "/")
	if len(s) != 2 {
		return "", raw
	}
	return s[0], s[1]
}

// FormatParent takes in a parent ("a") and result name ("b")
func FormatParent(cluster, namespace string) string {
	return fmt.Sprintf("clusters/%s/namespaces/%s", cluster, namespace)
}

// FormatName takes in a parent ("a") and result name ("b") and
// returns the full resource name ("a/results/b").
func FormatName(parent, name string) string {
	return fmt.Sprintf("%s/results/%s", parent, name)
}

// FormatParentDB takes in a parent ("a") and result name ("b")
func FormatParentDB(cluster, namespace string) string {
	return fmt.Sprintf("%s/%s", cluster, namespace)
}

// ToStorage converts an API Result into its corresponding database storage
// equivalent.
// parent,name should be the name parts (e.g. not containing "/results/").
func ToStorage(r *rpb.Result) (*models.Result, error) {
	cluster, namespace, name, err := ParseName(r.GetName())
	if err != nil {
		return nil, err
	}

	parent := FormatParentDB(cluster, namespace)

	id := r.GetId()

	result := &models.Result{
		Parent:      parent,
		ID:          id,
		Name:        name,
		Annotations: r.Annotations,
		Etag:        r.Etag,
	}

	if r.CreateTime.IsValid() {
		result.CreatedTime = r.CreateTime.AsTime()
	}
	if r.UpdateTime.IsValid() {
		result.UpdatedTime = r.UpdateTime.AsTime()
	}

	if s := r.GetSummary(); s != nil {
		if s.GetRecord() == "" || s.GetType() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "record and type fields required for RecordSummary")
		}
		if !record.NameRegex.MatchString(s.GetRecord()) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid record format")
		}
		if err := record.ValidateType(s.GetType()); err != nil {
			return nil, err
		}

		summary := models.RecordSummary{
			Record:      s.GetRecord(),
			Type:        s.GetType(),
			Status:      int32(s.GetStatus()),
			Annotations: s.Annotations,
		}
		if s.StartTime.IsValid() {
			summary.StartTime = ptr.Time(s.StartTime.AsTime())
		}
		if s.EndTime.IsValid() {
			summary.EndTime = ptr.Time(s.EndTime.AsTime())
		}
		result.Summary = summary
	}
	return result, nil
}

// ToAPI converts a database storage Result into its corresponding API
// equivalent.
func ToAPI(r *models.Result) *rpb.Result {
	var summary *rpb.RecordSummary
	if r.Summary.Record != "" {
		summary = &rpb.RecordSummary{
			Record:      r.Summary.Record,
			Type:        r.Summary.Type,
			StartTime:   newTS(r.Summary.StartTime),
			EndTime:     newTS(r.Summary.EndTime),
			Status:      rpb.RecordSummary_Status(r.Summary.Status),
			Annotations: r.Summary.Annotations,
		}
	}

	cluster, namespace := ParseParentDB(r.Parent)

	return &rpb.Result{
		Name:        FormatName(FormatParent(cluster, namespace), r.Name),
		Id:          r.ID,
		CreateTime:  timestamppb.New(r.CreatedTime),
		UpdateTime:  timestamppb.New(r.UpdatedTime),
		Annotations: r.Annotations,
		Etag:        r.Etag,
		Summary:     summary,
	}
}

func newTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// Match determines whether the given CEL filter matches the result.
func Match(r *rpb.Result, prg cel.Program) (bool, error) {
	if r == nil {
		return false, nil
	}
	return celenv.Match(prg, map[string]interface{}{
		"result": r,
	})
}

// UpdateEtag updates the etag field of a result according to its content.
// The result should at least have its `Id` and `UpdatedTime` fields set.
func UpdateEtag(r *models.Result) error {
	if r.ID == "" {
		return fmt.Errorf("the ID field must be set")
	}
	if r.UpdatedTime.IsZero() {
		return status.Error(codes.Internal, "the UpdatedTime field must be set")
	}
	r.Etag = fmt.Sprintf("%s-%v", r.ID, r.UpdatedTime.UnixNano())
	return nil
}

// GetFilteredPaginatedSortedResults returns the specified number of results that
// match the given CEL program.
func GetFilteredPaginatedSortedResults(ctx context.Context, db *gorm.DB, parent string, start string, pageSize int, prg cel.Program, sortOrder string) ([]*rpb.Result, error) {
	out := make([]*rpb.Result, 0, pageSize)
	batcher := pagination.NewBatcher(pageSize)

	for len(out) < pageSize {
		batchSize := batcher.Next()
		dbresults := make([]*models.Result, 0, batchSize)
		q := db.WithContext(ctx).Where("parent = ? AND id > ?", parent, start)
		if sortOrder != "" {
			q.Order(sortOrder)
		}
		q.Limit(batchSize).Find(&dbresults)
		if err := errors.Wrap(q.Error); err != nil {
			return nil, err
		}

		// Only return results that match the filter.
		for _, r := range dbresults {
			api := ToAPI(r)
			ok, err := Match(api, prg)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}

			out = append(out, api)
			if len(out) >= pageSize {
				return out, nil
			}
		}

		// We fetched less results than requested - this means we've exhausted
		// all items.
		if len(dbresults) < batchSize {
			break
		}

		// Set params for next batch.
		start = dbresults[len(dbresults)-1].ID
		batcher.Update(len(dbresults), batchSize)
	}
	return out, nil
}

func GetResult(ctx context.Context, db *gorm.DB, parent, name string) (*models.Result, error) {
	r := &models.Result{}
	q := db.
		WithContext(ctx).
		Where(&models.Result{Parent: parent, Name: name}).First(r)
	if err := errors.Wrap(q.Error); err != nil {
		return nil, err
	}
	return r, nil
}

func GetResultID(ctx context.Context, db *gorm.DB, parent, resultName string) (string, error) {
	id := new(resultID)
	q := db.WithContext(ctx).
		Model(&models.Result{}).
		Where(&models.Result{Parent: parent, Name: resultName}).
		First(id)
	if err := errors.Wrap(q.Error); err != nil {
		return "", err
	}
	return id.ID, nil
}
