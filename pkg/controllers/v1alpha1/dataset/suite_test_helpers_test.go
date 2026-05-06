package dataset

import "testing"

func TestAllowSkippingEnvtestStartFailure(t *testing.T) {
	t.Run("empty value stays disabled", func(t *testing.T) {
		t.Setenv(skipEnvtestStartFailureEnvVar, "")

		if allowSkippingEnvtestStartFailure() {
			t.Fatalf("expected empty %s value to keep opt-in skip disabled", skipEnvtestStartFailureEnvVar)
		}
	})

	t.Run("accepts explicit opt in", func(t *testing.T) {
		t.Setenv(skipEnvtestStartFailureEnvVar, "true")

		if !allowSkippingEnvtestStartFailure() {
			t.Fatalf("expected %s=true to enable opt-in skip", skipEnvtestStartFailureEnvVar)
		}
	})

	t.Run("rejects invalid values", func(t *testing.T) {
		t.Setenv(skipEnvtestStartFailureEnvVar, "sometimes")

		if allowSkippingEnvtestStartFailure() {
			t.Fatalf("expected invalid %s value to keep opt-in skip disabled", skipEnvtestStartFailureEnvVar)
		}
	})
}
