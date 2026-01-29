package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GetEnvsDifference", func() {
	DescribeTable("should return correct env difference",
		func(base, filter, expected []corev1.EnvVar) {
			result := GetEnvsDifference(base, filter)
			Expect(result).To(ConsistOf(expected))
		},
		Entry("nil_envs",
			[]corev1.EnvVar{},
			[]corev1.EnvVar{},
			[]corev1.EnvVar{},
		),
		Entry("test base envs are nil",
			[]corev1.EnvVar{},
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
			[]corev1.EnvVar{},
		),
		Entry("test exclude envs are nil",
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
			[]corev1.EnvVar{},
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
		),
		Entry("test base envs are same with exclude envs",
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
			[]corev1.EnvVar{},
		),
		Entry("test base envs includes all exclude envs",
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				env3 := corev1.EnvVar{Name: "env3"}
				env4 := corev1.EnvVar{Name: "env4"}
				return []corev1.EnvVar{env1, env2, env3, env4}
			}(),
			func() []corev1.EnvVar {
				env2 := corev1.EnvVar{Name: "env2"}
				env4 := corev1.EnvVar{Name: "env4"}
				return []corev1.EnvVar{env2, env4}
			}(),
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env3 := corev1.EnvVar{Name: "env3"}
				return []corev1.EnvVar{env1, env3}
			}(),
		),
		Entry("test base envs do not include with exclude envs",
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
			func() []corev1.EnvVar {
				env3 := corev1.EnvVar{Name: "env3"}
				env4 := corev1.EnvVar{Name: "env4"}
				return []corev1.EnvVar{env3, env4}
			}(),
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				return []corev1.EnvVar{env1, env2}
			}(),
		),
		Entry("test base envs includes partial exclude envs",
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env2 := corev1.EnvVar{Name: "env2"}
				env3 := corev1.EnvVar{Name: "env3"}
				return []corev1.EnvVar{env1, env2, env3}
			}(),
			func() []corev1.EnvVar {
				env2 := corev1.EnvVar{Name: "env2"}
				env4 := corev1.EnvVar{Name: "env4"}
				return []corev1.EnvVar{env2, env4}
			}(),
			func() []corev1.EnvVar {
				env1 := corev1.EnvVar{Name: "env1"}
				env3 := corev1.EnvVar{Name: "env3"}
				return []corev1.EnvVar{env1, env3}
			}(),
		),
	)
})
