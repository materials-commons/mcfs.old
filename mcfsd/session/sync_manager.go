package session

import (
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/materials-commons/mcfs/mcerr"
)

// ProjectSyncState tracks the current sync state for a project.
type ProjectSyncState struct {
	ProjectID   string        // Project being synced
	User        string        // User doing sync
	Started     time.Time     // Time sync started
	Uploaded    int64         // Bytes uploaded
	LastItem    string        // Last item synced
	TokenID     string        // Sync Token
	SawActivity bool          // Was there any activity on the sync?
	Expires     time.Duration // When does the sync token expire
}

type syncSessionManager struct {
	mutex    sync.RWMutex                 // Protect access to mutable data
	projects map[string]*ProjectSyncState // Projects currently being sync
}

var syncSession = newSyncSessionManager()

func newSyncSessionManager() *syncSessionManager {
	return &syncSessionManager{projects: make(map[string]*ProjectSyncState)}
}

func (s *syncSessionManager) acquireSyncToken(user, project string) (string, error) {
	defer s.mutex.Unlock()
	s.mutex.Lock()

	projState := s.projects[project]
	if projState != nil {
		return "", mcerr.ErrInUse
	}

	projState = &ProjectSyncState{
		ProjectID: project,
		User:      user,
		TokenID:   uuid.NewRandom().String(),
		//Expires:
	}

	return "", nil
}

func (s *syncSessionManager) expireSyncSession(project *ProjectSyncState) {
	for {
		project.expireSync()
		s.mutex.Lock()
		if !project.SawActivity {
			break
		}

		// If we are here then there was sync activity between the
		// expiration and the time we acquired the lock. So release
		// lock and start waiting again.
		s.mutex.Unlock()

	}
	// Perform cleanup and then release the lock.
	s.mutex.Unlock()

}

func (p *ProjectSyncState) expireSync() {
LOOP:
	for {
		select {
		case <-time.After(15 * time.Second):
			if !p.SawActivity {
				break LOOP
			}
			p.SawActivity = false
		}
	}
}
