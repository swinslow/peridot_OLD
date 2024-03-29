// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

import (
	"fmt"
)

// DoReset is the function for JobReset, and partially resets peridot by
// dropping all DB tables.
func (co *Coordinator) DoReset() error {
	// remove all DB tables
	err := co.db.ResetDB()
	if err != nil {
		return fmt.Errorf("couldn't delete all DB tables: %v", err)
	}

	return nil
}
