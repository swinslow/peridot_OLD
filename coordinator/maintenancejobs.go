// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

import (
    "fmt"
)

func (co *Coordinator) DoReset() error {
    // remove all DB tables
    err := co.db.ResetDB()
    if err != nil {
        return fmt.Errorf("couldn't delete all DB tables: %v")
    }

    return nil
}