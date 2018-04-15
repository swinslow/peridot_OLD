// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

import (
	"fmt"

	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/database"
	"github.com/swinslow/peridot/repomanager"
)

type Coordinator struct {
	rm  *repomanager.RepoManager
	db  *database.DB
	cfg *config.Config
}

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

	return nil
}
