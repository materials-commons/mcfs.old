package servers

import (
	"github.com/materials-commons/mcfs/server/servers/access"
)

// Maps each server instance to name.
var servers = map[string]Server{
	"Access": {Instance: access.Server()},
}

// Start starts all server instances.
func Start() {
	for _, s := range servers {
		s.Start()
	}
}

// StartNamed starts the named server instances.
func StartNamed(serverNames ...string) {
	for _, serverName := range serverNames {
		s, found := servers[serverName]
		if found {
			s.Start()
		}
	}
}

// Stop stops all server instances.
func Stop() {
	for _, s := range servers {
		s.Stop()
	}
}

// StopNamed stops the named server instances.
func StopNamed(serverNames ...string) {
	for _, serverName := range serverNames {
		s, found := servers[serverName]
		if found {
			s.Stop()
		}
	}
}
