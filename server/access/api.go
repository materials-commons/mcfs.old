package access

import (
	"github.com/materials-commons/base/schema"
)

// GetUserByAPIKey returns the User for a given APIKey.
func GetUserByAPIKey(apikey string) (*schema.User, error) {
	request := request{
		command: acGetUser,
		arg:     apikey,
	}
	server.request <- request
	response := <-server.response
	return response.user, response.err
}
