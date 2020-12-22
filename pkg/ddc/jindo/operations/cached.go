package operations

import "fmt"

// clean cache with a preset timeout of 60s
func (a JindoFileUtils) CleanCache(path string) (err error) {
	var (
		// TODO Jindofs clean cache command
		command = []string{"timeout", "-t", "60", "hadoop", "fs", "free", "-f", path}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}
