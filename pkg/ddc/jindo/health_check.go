package jindo

import "fmt"

func (e *JindoEngine) CheckRuntimeHealthy() (err error) {
	ready, err := e.CheckWorkersReady()
	if err != nil {
		e.Log.Error(err, "Check if the workers are ready")
		return err
	}

	if !ready {
		err = fmt.Errorf("the workers %s in %s are not healthy",
			e.getWorkertName(),
			e.namespace)
	}

	return
}
