package fileprefetcher

// Environment variables for file prefetcher
const (
	envKeyFilePrefetcherFileList       = "FILE_PREFETCHER_FILE_LIST"
	envKeyFilePrefetcherAsyncPrefetch  = "FILE_PREFETCHER_ASYNC_PREFETCH"
	envKeyFilePrefetcherTimeoutSeconds = "FILE_PREFETCHER_TIMEOUT_SECONDS"

	envKeyFilePrefetcherImage = "FILE_PREFETCHER_IMAGE"
)

// Constants for file prefetcher
const (
	filePrefetcherContainerName         = "fluid-file-prefetcher"
	filePrefetcherStatusVolumeName      = "fluid-file-prefetcher-status-vol"
	filePrefetcherStatusVolumeMountPath = "/tmp/fluid-file-prefetcher/status"

	filePrefetcherDefaultFileList          = "<ALL>"
	filePrefetcherDefaultTimeoutSecondsStr = "120"
)
