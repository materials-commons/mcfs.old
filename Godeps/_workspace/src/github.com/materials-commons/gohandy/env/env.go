package env

import (
	"os"
)

// GetDefault retrieves the value of the environment variable named by the key. If the value is empty, then it
// returns defaultValue, otherwise it returns the value of the variable.
func GetDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val
}

// GetenvDefaultSet retrieves the value of the environment variable named by key. If the value is empty, then it
// attempts to set key to defaultValue, and returns defaultValue. If key was already set it just returns the
// that value, and doesn't set it.. It returns an error if it was not able to set it.
func GetDefaultSet(key, defaultValue string) (string, error) {
	val := os.Getenv(key)
	switch {
	case val == "":
		// If val is "" then we attempt to set the key to
		// the defaultValue. If that fails return an error.
		if err := os.Setenv(key, defaultValue); err != nil {
			return val, err
		}
		return defaultValue, nil
	default:
		// Key had a value, return it.
		return val, nil
	}
}
