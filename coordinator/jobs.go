// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

type JobType int

const (
	// No-op
	JobNop JobType = iota

	// ===== Repo Management =====

	// First retrieval (clone) of files for a new repo
	JobCloneRepo

	// Update an existing, previously-cloned repo
	JobUpdateRepo

	// After a first clone or later update, add the directories and files
	// to the database and hash manager
	JobPrepareFiles

	// ===== Maintenance =====

	// Reset -- drop all tables
	// USE CAUTION before calling this!
	JobReset

	// TO DO: Scan through database, repos and hash manager and check for
	// inconsistencies
)
