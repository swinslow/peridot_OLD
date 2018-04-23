// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

// JobType represents a job that gets called for a Coordinator.
type JobType int

const (
	// JobNop signifies a no-op job
	JobNop JobType = iota

	// ===== Repo Management =====

	// JobCloneRepo signifies a job for the first retrieval (clone) of files
	// for a new repo
	JobCloneRepo

	// JobUpdateRepo signifies a job to update an existing, previously-cloned
	// repo
	JobUpdateRepo

	// JobPrepareFiles signifies a job that gets called after a first clone
	// or a later update, to add the directories and files to the database
	// and hash manager
	JobPrepareFiles

	// ===== Maintenance =====

	// JobReset signifies a job that is called to partially reset peridot by
	// dropping all DB tables; USE CAUTION before calling this!
	JobReset

	// TO DO: Scan through database, repos and hash manager and check for
	// inconsistencies
)
