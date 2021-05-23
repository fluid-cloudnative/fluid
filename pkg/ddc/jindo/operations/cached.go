package operations

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		if utils.IgnoreNotFound(err) == nil {
			fmt.Printf("Failed to clean cache due to %v", err)
			return nil
		}
		return
	} else {
		time.Sleep(30 * time.Second)
	}

	return
}
