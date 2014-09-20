package autoupdate

import (
	"github.com/materials-commons/mcfs/materialsd/config"
)

// A Updater keeps track of the status of binary and website updates
// and downloads updates when they are avaiable.
type Updater struct {
	downloaded    string
	binaryUpdated bool
}

// NewUpdater creates a new Updater instance.
func NewUpdater() *Updater {
	return &Updater{
		downloaded:    "",
		binaryUpdated: false,
	}
}

// UpdatesAvailable checks if updates are available for either the website
// or the materials binary. If updates are available it will download them.
func (u *Updater) UpdatesAvailable() bool {
	updateAvailable := false

	if materials.Update(config.Config.MaterialsCommons.Download) {
		updateAvailable = true
		u.binaryUpdated = true
	}

	return updateAvailable
}

// ApplyUpdates deploys updates that have been downloaded. If the materials
// binary has been updated then it restarts the server.
func (u *Updater) ApplyUpdates() {
	if u.binaryUpdated {
		materials.Restart()
	}
}

// BinaryUpdate returns true if the materials binary has been updated.
func (u *Updater) BinaryUpdate() bool {
	return u.binaryUpdated
}
