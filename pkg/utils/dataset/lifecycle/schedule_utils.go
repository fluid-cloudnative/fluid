package lifecycle

import "k8s.io/api/core/v1"

// toleratesTaints tolerates the taints in node
func toleratesTaints(taints []v1.Taint, tolerations []v1.Toleration) bool {
	filteredTaints := []v1.Taint{}
	for _, taint := range taints {
		if taint.Effect == v1.TaintEffectNoExecute || taint.Effect == v1.TaintEffectNoSchedule {
			filteredTaints = append(filteredTaints, taint)
		}
	}

	if len(tolerations) == 0 {
		return false
	}

	toleratesTaint := func(taint v1.Taint) bool {
		for _, toleration := range tolerations {
			if toleration.ToleratesTaint(&taint) {
				return true
			}
		}

		return false
	}

	for _, taint := range filteredTaints {
		if !toleratesTaint(taint) {
			return false
		}
	}

	return true
}
