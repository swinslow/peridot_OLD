// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

import (
	"fmt"

	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/database"
	"github.com/swinslow/peridot/hashmanager"
	"github.com/swinslow/peridot/repomanager"
)

// Coordinator holds the data objects used to manage various of
// peridot's activities.
type Coordinator struct {
	rm  *repomanager.RepoManager
	hm  *hashmanager.HashManager
	db  *database.DB
	cfg *config.Config
}

// Prepare is called with existing Config and Database objects and creates the
// other maanger objects for the Coordinator.
func (co *Coordinator) Prepare(cfg *config.Config, db *database.DB) error {
	var err error

	if co == nil {
		return fmt.Errorf("must pass non-nil Coordinator to Prepare()")
	}
	if cfg == nil {
		return fmt.Errorf("must pass config string to Prepare()")
	}
	if db == nil {
		return fmt.Errorf("must pass valid database to Prepare()")
	}

	co.db = db

	co.rm = &repomanager.RepoManager{}
	err = co.rm.PrepareRM(cfg, co.db)
	if err != nil {
		return err
	}

	co.hm = &hashmanager.HashManager{}
	err = co.hm.PrepareHM(cfg, co.db)
	if err != nil {
		return err
	}

	return nil
}
