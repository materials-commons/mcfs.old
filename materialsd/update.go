package materials

import (
	"bitbucket.org/kardianos/osext"
	"fmt"
	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/gohandy/file"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Maps the given os to the binary name. Only
// windows adds an extension to the name.
var executable = map[string]string{
	"windows": "materials.exe",
	"darwin":  "materials",
	"linux":   "materials",
}

/*
var bgcommands = map[string][]string{
	"windows": []string{"start", "/min"},
	"darwin":  []string{"nohup"},
	"linux":   []string{"nohup"},
}
*/

type restartFunction func(commandPath string)

var restartFunc = map[string]restartFunction{
	"windows": restartWindows,
	"darwin":  restartDarwin,
	"linux":   restartLinux,
}

// Restart restarts the materials command. It starts a new command
// and then exits the current command. The new command is started
// using nohup so the parent exiting doesn't terminate it. The new
// command also does a retry on the port to give the parent some
// time to exit and release it.
func Restart() {
	commandPath, err := osext.Executable()
	if err != nil {
		fmt.Printf("Unable to determine my executable path: %s\n", err.Error())
		return
	}

	f, ok := restartFunc[runtime.GOOS]
	if !ok {
		panic(fmt.Sprintf("Don't know what the restart command is on %s platform", runtime.GOOS))
	}
	f(commandPath)
	/*
		run, ok := bgcommands[runtime.GOOS]
		if !ok {
			panic(fmt.Sprintf("Don't know what the background command is on %s platform", runtime.GOOS))
		}
		run = append(run, commandPath, "--server", "--retry=10")
		command := exec.Command(run[0], run[1:]...)
		command.Start()
		os.Exit(0)
	*/
}

// Update replaces the current binary with a new one if they are different.
// It determines if they are different by comparing their checksum's. Update
// downloads the binary at the specified url. It modifies the url to include
// the os type in the path. This is determined by using runtime.GOOS.
func Update(url string) bool {
	myPath, myChecksum := me()
	downloadedPath, downloadedChecksum, err := downloaded(url)

	switch {
	case err != nil:
		return false
	case myChecksum != downloadedChecksum:
		replaceMe(myPath, downloadedPath)
		return true
	default:
		return false
	}
}

// me returns current binary path and checksum.
func me() (mypath string, mychecksum uint32) {
	mypath, _ = osext.Executable()
	mychecksum = file.Checksum32(mypath)
	return
}

// downloaded downloads a new binary and returns its path, checksum and
// an error condition. The error is nil if no error occurred.
func downloaded(url string) (dlpath string, dlchecksum uint32, err error) {
	dlchecksum = 0
	dlpath, err = downloadNewBinary(binaryURL(url))
	if err != nil {
		return
	}

	dlchecksum = file.Checksum32(dlpath)
	return
}

// binaryUrl returns the url to download the binary for the current OS.
func binaryURL(url string) string {
	return binaryURLForRuntime(url, runtime.GOOS)
}

// binaryUrlForRuntime returns the url to download the binary for a given OS.
func binaryURLForRuntime(url, whichRuntime string) string {
	exe, ok := executable[whichRuntime]
	if !ok {
		panic(fmt.Sprintf("Unknown runtime: %s", whichRuntime))
	}
	s := []string{url, whichRuntime, exe}
	return strings.Join(s, "/")
}

// downloadNewBinary downloads our binary from the given url.
// It determines the correct name to save the binary to and
// saves it into the OS tempdir.
func downloadNewBinary(url string) (string, error) {
	client := ezhttp.NewInsecureClient()
	executable, _ := osext.Executable()
	materialsName := filepath.Base(executable)
	path := filepath.Join(os.TempDir(), materialsName)
	status, err := client.FileGet(url, path)
	switch {
	case err != nil:
		return "", err
	case status != 200:
		return "", fmt.Errorf("unable to download file, status code %d", status)
	default:
		return path, nil
	}
}

// Replaces current binary with the downloaded one.
func replaceMe(mypath, downloadedPath string) error {
	return os.Rename(downloadedPath, mypath)
}

func restartWindows(commandPath string) {
	run := []string{"start", "/min", commandPath, "--server", "--retry=10"}
	command := exec.Command(run[0], run[1:]...)
	command.Start()
	os.Exit(0)
}

func restartDarwin(commandPath string) {
	/*
		restartCommand := filepath.Join(filepath.Dir(commandPath), "materials-restart-darwin")
		command := exec.Command("nohup", restartCommand)
		command.Start()
	*/
	os.Exit(0)
}

func restartLinux(commandPath string) {
	run := []string{"nohup", commandPath, "--server", "--retry=10"}
	command := exec.Command(run[0], run[1:]...)
	command.Start()
	os.Exit(0)
}
