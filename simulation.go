// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"github.com/simagix/keyhole/sim"
)

// StartSimulation kicks off simulation
func StartSimulation(runner *sim.Runner) error {
	var err error
	if err = runner.Start(); err != nil {
		return err
	}
	return runner.CollectAllStatus()
}
