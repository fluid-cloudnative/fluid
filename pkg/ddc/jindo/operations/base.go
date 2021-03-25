package operations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
)

type JindoFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewJindoFileUtils(podName string, containerName string, namespace string, log logr.Logger) JindoFileUtils {

	return JindoFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// exec with timeout
func (a JindoFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1500)
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = a.execWithoutTimeout(command, verbose)
		ch <- "done"
	}()

	select {
	case <-ch:
		a.log.V(1).Info("execute in time", "command", command)
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}

	return
}

// execWithoutTimeout
func (a JindoFileUtils) execWithoutTimeout(command []string, verbose bool) (stdout string, stderr string, err error) {
	stdout, stderr, err = kubeclient.ExecCommandInContainer(a.podName, a.container, a.namespace, command)
	if err != nil {
		a.log.Info("Stdout", "Command", command, "Stdout", stdout)
		a.log.Error(err, "Failed", "Command", command, "FailedReason", stderr)
		return
	}
	if verbose {
		a.log.Info("Stdout", "Command", command, "Stdout", stdout)
	}

	return
}

// Get summary info of the Alluxio Engine
func (a JindoFileUtils) ReportSummary() (summary string, err error) {
	var (
		command = []string{"/sdk/bin/jindo", "jfs", "-report"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return stdout, err
	}
	return stdout, err
}

func (a JindoFileUtils) GetUfsTotalSize(url string) (summary string, err error) {
	var (
		command = []string{"hadoop", "fs", "-count", url}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.execWithoutTimeout(command, false)

	str := strings.Fields(stdout)

	if len(str) < 3 {
		err = fmt.Errorf("failed to parse %s in Count method", str)
		return
	}

	stdout = str[2]

	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return stdout, err
	}
	return stdout, err
}
