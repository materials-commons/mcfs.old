package access

import (
	"github.com/materials-commons/mcfs/common/schema"
)

// GetUserByAPIKey returns the User for a given APIKey.
func GetUserByAPIKey(apikey string) (*schema.User, error) {
	request := request{
		command: acGetUser,
		arg:     apikey,
	}

	if err := server.Send(&request); err != nil {
		return nil, err
	}

	response := server.Recv()
	return response.user, response.err
}
