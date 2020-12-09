package operations

import (
	"context"
	"fmt"
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

// Get summary info of the Alluxio Engine
func (a JindoFileUtils) ReportSummary() (summary string, err error) {
	var (
		// TODO report summary
		command = []string{"alluxio", "fsadmin", "report", "summary"}
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

