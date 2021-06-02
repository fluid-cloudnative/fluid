package common

const (
	WebhookName        = "fluid-pod-admission-webhook"
	WebhookServiceName = "fluid-pod-admission-webhook"
	WebhookServicePath = "/mutate-v1-pod"

	// this file comes from tools/certificate.sh
	CertificationGenerateFile = "/usr/local/bin/certificate.sh"
)
