package session

import (
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/materials-commons/mcfs/mcerr"
)

// ProjectSyncState tracks the current sync state for a project.
type ProjectSyncState struct {
	mutex       sync.Mutex // Mutex to coordinate access to project state.
	ProjectID   string     // Project being synced
	User        string     // User doing sync
	Started     time.Time  // Time sync started
	Uploaded    int64      // Bytes uploaded
	LastItem    string     // Last item synced
	TokenID     string     // Sync Token
	SawActivity bool       // Was there any activity on the sync?
}

type syncSessionManager struct {
	mutex    sync.RWMutex                 // Protect access to mutable data
	projects map[string]*ProjectSyncState // Projects currently being sync
}

var syncSession = NewSyncSessionManager()

func NewSyncSessionManager() *syncSessionManager {
	return &syncSessionManager{projects: make(map[string]*ProjectSyncState)}
}

// AcquireSyncToken creates a new sync token for a project if that project doesn't
// currently have a sync token associated with it. It also launches a go routine
// for timeout on the token.
func (s *syncSessionManager) AcquireSyncToken(user, project string) (string, error) {
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
	}

	go s.expireSyncSession(projState)

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
	delete(s.projects, project.ProjectID)
	s.mutex.Unlock()

}

func (p *ProjectSyncState) expireSync() {
LOOP:
	for {
		// Start flag false. It will get set to true if there is activity in the specified
		// time period.
		p.mutex.Lock()
		p.SawActivity = false
		p.mutex.Unlock()
		select {
		case <-time.After(15 * time.Second):
			if !p.SawActivity {
				// No activity so break out of loop
				break LOOP
			}

		}
	}
}
