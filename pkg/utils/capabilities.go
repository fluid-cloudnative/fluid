package utils

import (
	corev1 "k8s.io/api/core/v1"
)

func TrimCapabilities(inputs []corev1.Capability, excludeNames []string) (outputs []corev1.Capability) {
	outputs = []corev1.Capability{}

outer:
	for _, in := range inputs {
		for _, excludeName := range excludeNames {
			if string(in) == excludeName {
				log.V(1).Info("Skip the capability", "capability", in, "excludeCapability", excludeName)
				continue outer
			}
		}

		outputs = append(outputs, in)
	}

	return
}
