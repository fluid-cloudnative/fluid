package operations

import (
	"fmt"
	"time"
)

// clean cache with a preset timeout of 60s
func (a JindoFileUtils) CleanCache() (err error) {
	var (
		// jindo jfs -formatCache -force
		command = []string{"jindo", "jfs", "-formatCache", "-force"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)

	time.Sleep(30 * time.Second)

	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}
