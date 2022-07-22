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

// Package record provides utilities for manipulating and validating Records.
package record

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/cel-go/checker/decls"
	"github.com/tektoncd/results/pkg/server/db/errors"
	"github.com/tektoncd/results/pkg/server/db/pagination"
	"gorm.io/gorm"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	celenv "github.com/tektoncd/results/pkg/server/cel"
	"github.com/tektoncd/results/pkg/server/db/models"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	typeSize = 768
)

var (
	// NameRegex matches valid name specs for a Result.
	NameRegex   = regexp.MustCompile("^clusters/([a-z0-9:_-]{1,63})/namespaces/([a-z0-9_-]{1,63})/results/([a-z0-9_-]{1,63})/records/([a-z0-9_-]{1,63})$")
	ParentRegex = regexp.MustCompile("^clusters/([a-z0-9:_-]{1,63})/namespaces/([a-z0-9_-]{1,63})/results/([a-z0-9_-]{1,63})$")
)

// ParseParent splits a top-level parent into its individual (cluster, namespace)
func ParseParent(raw string) (cluster, namespace, result string, err error) {
	s := ParentRegex.FindStringSubmatch(raw)
	if len(s) != 4 {
		return "", "", "", status.Errorf(codes.InvalidArgument, "parent must match %s", ParentRegex.String())
	}
	return s[1], s[2], s[3], nil
}

// ParseName splits a full Result name into its individual (parent, result, name)
// components.
func ParseName(raw string) (cluster, namespace, result, name string, err error) {
	s := NameRegex.FindStringSubmatch(raw)
	if len(s) != 5 {
		return "", "", "", "", status.Errorf(codes.InvalidArgument, "name must match %s", NameRegex.String())
	}
	return s[1], s[2], s[3], s[4], nil
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
func FormatParent(cluster, namespace, result string) string {
	return fmt.Sprintf("clusters/%s/namespaces/%s/results/%s", cluster, namespace, result)
}

// FormatName takes in a parent ("a/results/b") and record name ("c") and
// returns the full resource name ("a/results/b/records/c").
func FormatName(parent, name string) string {
	return fmt.Sprintf("%s/records/%s", parent, name)
}

// FormatParentDB takes in a parent ("a") and result name ("b")
func FormatParentDB(cluster, namespace string) string {
	return fmt.Sprintf("%s/%s", cluster, namespace)
}

// ToStorage converts an API Record into its corresponding database storage
// equivalent.
// parent,result,name should be the name parts (e.g. not containing "/results/" or "/records/").
func ToStorage(r *rpb.Record) (*models.Record, error) {
	cluster, namespace, resultName, name, err := ParseName(r.GetName())
	if err != nil {
		return nil, err
	}

	if err := validateData(r.GetData()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id := r.GetId()

	dbr := &models.Record{
		Parent:     FormatParentDB(cluster, namespace),
		ResultName: resultName,

		ID:   id,
		Name: name,

		Type: r.GetData().GetType(),
		Data: r.GetData().GetValue(),

		Etag: r.Etag,
	}

	if r.CreateTime.IsValid() {
		dbr.CreatedTime = r.CreateTime.AsTime()
	}
	if r.UpdateTime.IsValid() {
		dbr.UpdatedTime = r.UpdateTime.AsTime()
	}

	return dbr, nil
}

// ToAPI converts a database storage Record into its corresponding API
// equivalent.
func ToAPI(r *models.Record) (*rpb.Record, error) {
	cluster, namespace := ParseParentDB(r.Parent)
	out := &rpb.Record{
		Name: FormatName(FormatParent(cluster, namespace, r.ResultName), r.Name),
		Id:   r.ID,
		Etag: r.Etag,
	}

	if !r.CreatedTime.IsZero() {
		out.CreateTime = timestamppb.New(r.CreatedTime)
	}
	if !r.UpdatedTime.IsZero() {
		out.UpdateTime = timestamppb.New(r.UpdatedTime)
	}

	if r.Data != nil {
		out.Data = &rpb.Any{
			Type:  r.Type,
			Value: r.Data,
		}
	}
	return out, nil
}

// Match determines whether the given CEL filter matches the result.
func Match(r *rpb.Record, prg cel.Program) (bool, error) {
	if r == nil {
		return false, nil
	}

	var m map[string]interface{}
	if d := r.GetData().GetValue(); d != nil {
		if err := json.Unmarshal(r.GetData().GetValue(), &m); err != nil {
			return false, err
		}
	}

	return celenv.Match(prg, map[string]interface{}{
		"name":      r.GetName(),
		"data_type": r.GetData().GetType(),
		"data":      m,
	})
}

// UpdateEtag updates the etag field of a record according to its content.
// The record should at least have its `Id` and `UpdatedTime` fields set.
func UpdateEtag(r *models.Record) error {
	if r.ID == "" {
		return fmt.Errorf("the ID field must be set")
	}
	if r.UpdatedTime.IsZero() {
		return status.Error(codes.Internal, "the UpdatedTime field must be set")
	}
	r.Etag = fmt.Sprintf("%s-%v", r.ID, r.UpdatedTime.UnixNano())
	return nil
}

func validateData(m *rpb.Any) error {
	if err := ValidateType(m.GetType()); err != nil {
		return err
	}

	if m == nil {
		return nil
	}
	switch m.GetType() {
	case "pipeline.tekton.dev/TaskRun":
		return json.Unmarshal(m.GetValue(), &v1beta1.TaskRun{})
	case "pipeline.tekton.dev/PipelineRun":
		return json.Unmarshal(m.GetValue(), &v1beta1.PipelineRun{})
	default:
		// If it's not a well known type, just check that the message is a valid JSON document.
		return json.Unmarshal(m.GetValue(), &json.RawMessage{})
	}
}

func ValidateType(t string) error {
	// Certain DBs like sqlite will massage CHAR types to TEXT, so enforce
	// this in our code for consistency.
	if len(t) > typeSize {
		return status.Errorf(codes.InvalidArgument, "type must not exceed %d characters", typeSize)
	}
	return nil
}

// GetRecord returns a specified record from database
func GetRecord(ctx context.Context, db *gorm.DB, parent, result, name string) (*models.Record, error) {
	r := &models.Record{}
	q := db.
		WithContext(ctx).
		Where(&models.Record{Result: models.Result{Parent: parent, Name: result}, Name: name}).
		First(r)
	if err := errors.Wrap(q.Error); err != nil {
		return nil, err
	}
	return r, nil
}

// GetFilteredPaginatedSortedRecords returns the specified number of results that match the given CEL program.
func GetFilteredPaginatedSortedRecords(ctx context.Context, db *gorm.DB, parent, result, start string, pageSize int, prg cel.Program, sortOrder string) ([]*rpb.Record, error) {
	out := make([]*rpb.Record, 0, pageSize)
	batcher := pagination.NewBatcher(pageSize)
	for len(out) < pageSize {
		batchSize := batcher.Next()
		dbrecords := make([]*models.Record, 0, batchSize)
		q := db.WithContext(ctx).Where("parent = ? AND id > ?", parent, start)
		// Specifying `-` allows users to read Records across Results.
		// See https://google.aip.dev/159 for more details.
		if result != "-" {
			q = q.Where("result_name = ?", result)
		}
		if sortOrder != "" {
			q = q.Order(sortOrder)
		}
		q = q.Limit(batchSize).Find(&dbrecords)
		if err := errors.Wrap(q.Error); err != nil {
			return nil, err
		}

		// Only return results that match the filter.
		for _, r := range dbrecords {
			api, err := ToAPI(r)
			if err != nil {
				return nil, err
			}
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
		if len(dbrecords) < batchSize {
			break
		}

		// Set params for next batch.
		start = dbrecords[len(dbrecords)-1].ID
		batcher.Update(len(dbrecords), batchSize)
	}
	return out, nil
}

// RecordCEL defines the CEL environment for querying Record data.
// Fields are broken up explicitly in order to support dynamic handling of the
// data field as a key-value document.
func RecordCEL() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Types(&rpb.Record{}),
		cel.Declarations(decls.NewVar("name", decls.String)),
		cel.Declarations(decls.NewVar("data_type", decls.String)),
		cel.Declarations(decls.NewVar("data", decls.Dyn)),
	)
}
