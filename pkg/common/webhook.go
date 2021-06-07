package common

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	WebhookName            = "fluid-pod-admission-webhook"
	WebhookServiceName     = "fluid-pod-admission-webhook"
	WebhookSchedulePodPath = "mutate-fluid-io-v1alpha1-schedulepod"

	// this file comes from tools/certificate.sh
	CertificationGenerateFile = "/usr/local/bin/certificate.sh"
)

// AdmissionHandler wrappers admission.Handler, but adding client-go capablities
type AdmissionHandler interface {
	admission.Handler

	Setup(client client.Client)
}
