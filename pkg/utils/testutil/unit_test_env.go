package testutil

import "os"

const FluidUnitTestEnv = "FLUID_UNIT_TEST"

func IsUnitTest() bool {
	_, exists := os.LookupEnv(FluidUnitTestEnv)
	return exists
}
