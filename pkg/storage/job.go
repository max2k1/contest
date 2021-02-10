// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"github.com/facebookincubator/contest/pkg/job"
	"github.com/facebookincubator/contest/pkg/types"
)

// JobStorage defines the interface that implements persistence for job
// related information
type JobStorage interface {
	// Job request interface
	StoreJobRequest(request *job.Request) (types.JobID, error)
	GetJobRequest(jobID types.JobID) (*job.Request, error)

	// Job report interface
	StoreJobReport(report *job.JobReport) error
	GetJobReport(jobID types.JobID) (*job.JobReport, error)

	// Job enumeration interface
	ListJobs(query *JobQuery) ([]types.JobID, error)
}

// JobStorageManager implements JobStorage interface
type JobStorageManager struct {
}

// StoreJobRequest submits a job request to the storage layer
func (jsm JobStorageManager) StoreJobRequest(request *job.Request) (types.JobID, error) {
	return storage.StoreJobRequest(request)
}

// GetJobRequest fetches a job request from the storage layer
func (jsm JobStorageManager) GetJobRequest(jobID types.JobID) (*job.Request, error) {
	return storage.GetJobRequest(jobID)
}

// StoreJobReport submits a job report to the storage layer
func (jsm JobStorageManager) StoreJobReport(report *job.JobReport) error {
	return storage.StoreJobReport(report)
}

// GetJobReport fetches a job report from the storage layer
func (jsm JobStorageManager) GetJobReport(jobID types.JobID) (*job.JobReport, error) {
	return storage.GetJobReport(jobID)
}

// ListJobs returns list of job IDs matching the query
func (jsm JobStorageManager) ListJobs(query *JobQuery) ([]types.JobID, error) {
	return storage.ListJobs(query)
}

// NewJobStorageManager creates a new JobStorageManager object
func NewJobStorageManager() JobStorageManager {
	return JobStorageManager{}
}
