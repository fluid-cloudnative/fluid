package common

// Runtime for Vineyard
const (
	VineyardRuntime = "vineyard"

	VineyardMountType = "vineyard-fuse"

	VineyardChart = VineyardRuntime

	VineyardFuseIsGlobal = true

	DefaultVineyardMasterImage = "bitnami/etcd:3.5.10"

	DefaultVineyardWorkerImage = "vineyardcloudnative/vineyardd:latest"

	DefultVineyardFuseImage = "vineyardcloudnative/mount-vineyard-socket:latest"

	VineyardEngineImpl = VineyardRuntime
)

var (
	VineyardFuseNodeSelector = map[string]string{}
)
