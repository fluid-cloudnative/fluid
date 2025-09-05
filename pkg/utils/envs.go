package utils

import corev1 "k8s.io/api/core/v1"

// GetEnvsDifference calculates the difference between two env slices,
// returning envs that exist in the envs1 slice but not in the envs2 slice.
func GetEnvsDifference(envs1, envs2 []corev1.EnvVar) []corev1.EnvVar {
	retEnvs := make([]corev1.EnvVar, 0)
	envsRecord := make(map[string]bool)
	for _, env := range envs2 {
		envsRecord[env.Name] = true
	}
	for _, env := range envs1 {
		if !envsRecord[env.Name] {
			retEnvs = append(retEnvs, env)
		}
	}
	return retEnvs
}
