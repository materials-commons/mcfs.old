package config

import (
	"encoding/json"
	"github.com/materials-commons/mcfs/client/user"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// MaterialsCommonsConfig holds all the configuration information
// for accessing Materials Commons services.
type MaterialsCommonsConfig struct {
	API        string
	URL        string
	Download   string
	UploadHost string
	UploadPort int
}

// ServerConfig holds all the configuration for this server.
type ServerConfig struct {
	Port                uint
	SocketIOPort        uint
	Address             string
	Webdir              string
	UpdateCheckInterval time.Duration
	LastUpdateCheck     string
	NextUpdateCheck     string
	LastWebsiteUpdate   string
	LastServerUpdate    string
}

// UserConfig hold configuration for the user.
type UserConfig struct {
	*user.User
	DefaultProject string
}

// ConfigSettings holds all the individual configuration items.
type ConfigSettings struct {
	MaterialsCommons MaterialsCommonsConfig
	Server           ServerConfig
	User             UserConfig
}

type configFile map[string]interface{}

var defaultSettings = map[string]interface{}{
	"server_address":        "localhost",
	"server_port":           uint(8081),
	"socketio_port":         uint(8082),
	"update_check_interval": 4 * time.Hour,
	"MCFS_HOST":             "materialscommons.org",
	"MCFS_PORT":             35862,
	"MCURL":                 "https://materialscommons.org",
	"MCAPIURL":              "https://api.materialscommons.org",
	"MCDOWNLOADURL":         "https://download.materialscommons.org",
}

// Config is the single instance of the servers configuration settings.
var Config ConfigSettings

// EnvironmentVariables is a list of the environment variables the server looks for
// to override default settings.
var EnvironmentVariables = []string{
	"MATERIALS_PORT", "MATERIALS_ADDRESS", "MATERIALS_SOCKETIO_PORT",
	"MATERIALS_UPDATE_CHECK_INTERVAL", "MATERIALS_WEBDIR", "MCAPIURL",
	"MCURL", "MCDOWNLOADURL", "MCFS_HOST", "MCFS_PORT",
}

//*********************************************************
// TODO: Create an Initialize() for the materials package
// that encompasses all the other initialization, such
// as projects, and .user
//*********************************************************

// ConfigInitialize initializes the configuration.
func ConfigInitialize(user *user.User) {
	Config.User.User = user
	Config.setConfigOverrides()
}

func (c *ConfigSettings) setConfigOverrides() {
	configFromFile, _ := readConfigFile(c.User.DotMaterialsPath())
	c.Server.Port = getConfigUint("server_port", "MATERIALS_PORT", configFromFile)
	c.Server.Address = getConfigStr("server_address", "MATERIALS_ADDRESS", configFromFile)
	c.Server.SocketIOPort = getConfigUint("socketio_port", "MATERIALS_SOCKETIO_PORT", configFromFile)
	updateCheckInterval := getConfigDuration("update_check_interval", "MATERIALS_UPDATE_CHECK_INTERVAL", configFromFile)
	c.Server.UpdateCheckInterval = updateCheckInterval
	c.MaterialsCommons.API = getDefaultedConfigStr("MCAPIURL", "MCAPIURL")
	c.MaterialsCommons.URL = getDefaultedConfigStr("MCURL", "MCURL")
	c.MaterialsCommons.Download = getDefaultedConfigStr("MCDOWNLOADURL", "MCDOWNLOADURL")
	c.MaterialsCommons.UploadHost = getDefaultedConfigStr("MCFS_HOST", "MCFS_HOST")
	c.MaterialsCommons.UploadPort = getDefaultedConfigInt("MCFS_PORT", "MCFS_PORT")
	webdir := os.Getenv("MATERIALS_WEBDIR")
	if webdir == "" {
		webdir = filepath.Join(c.User.DotMaterialsPath(), "website")
	}

	c.Server.Webdir = webdir

	cf := configFromFile
	defaultProject, ok := cf["default_project"].(string)
	if ok {
		c.User.DefaultProject = defaultProject
	}
}

func getConfigUint(jsonName, envName string, c configFile) uint {
	envVal, err := strconv.ParseUint(os.Getenv(envName), 0, 32)
	jsonVal, ok := c[jsonName].(uint)

	switch {
	case err == nil:
		return uint(envVal)
	case ok && jsonVal != 0:
		return jsonVal
	default:
		val, _ := defaultSettings[jsonName].(uint)
		return val
	}
}

func getConfigDuration(jsonName, envName string, c configFile) time.Duration {
	envVal, err := strconv.ParseUint(os.Getenv(envName), 0, 32)
	jsonVal, ok := c[jsonName].(time.Duration)

	switch {
	case err == nil:
		return time.Duration(envVal)
	case ok && jsonVal != 0:
		return jsonVal
	default:
		val, _ := defaultSettings[jsonName].(time.Duration)
		return val
	}
}

func getConfigStr(jsonName, envName string, c configFile) string {
	envVal := os.Getenv(envName)
	jsonVal, ok := c[jsonName].(string)

	switch {
	case envVal != "":
		return envVal
	case ok && jsonVal != "":
		return jsonVal
	default:
		val, _ := defaultSettings[jsonName].(string)
		return val
	}
}

func getDefaultedConfigStr(envName, settingsName string) string {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return defaultSettings[settingsName].(string)
	}

	return envVal
}

func getDefaultedConfigInt(envName, settingsName string) int {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return defaultSettings[settingsName].(int)
	}

	i, _ := strconv.Atoi(envVal)
	return i
}

func readConfigFile(dotmaterialsPath string) (cf configFile, err error) {
	configPath := configPath(dotmaterialsPath)
	bytes, err := ioutil.ReadFile(configPath)
	var config configFile
	cf = config

	if err != nil {
		return cf, err
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		return cf, err
	}

	return config, nil
}

func configPath(path string) string {
	return filepath.Join(path, ".config")
}

func writeConfigFile(config configFile, dotmaterialsPath string) error {
	return nil
}

// APIURLPath constructs the url to access an api service. Includes the
// apikey. Prepends a "/" if needed.
func (c ConfigSettings) APIURLPath(service string) string {
	if string(service[0]) != "/" {
		service = "/" + service
	}
	uri := c.MaterialsCommons.API + service + "?apikey=" + c.User.APIKey
	return uri
}
