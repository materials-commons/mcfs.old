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

	if err := Send(&request); err != nil {
		return nil, err
	}

	response := Recv()
	return response.user, response.err
}