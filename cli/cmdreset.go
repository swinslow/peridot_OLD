// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/coordinator"
	"github.com/swinslow/peridot/database"
)

func CmdReset(co *coordinator.Coordinator, db *database.DB, cfg *config.Config) {
	err := co.DoReset()
	if err != nil {
		fmt.Printf("Error resetting: %v\n", err)
		return
	}

	fmt.Printf("Reset peridot. Did _not_ delete repos (do that manually).\n")
}
