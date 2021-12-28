package jindo

const (
	CSI_DRIVER = "fuse.csi.fluid.io"

	//fluid_PATH = "fluid_path"

	Mount_TYPE = "mount_type"

	SUMMARY_PREFIX_TOTAL_CAPACITY = "Total Capacity: "

	SUMMARY_PREFIX_USED_CAPACITY = "Used Capacity: "

	METADATA_SYNC_NOT_DONE_MSG = "[Calculating]"

	CHECK_METADATA_SYNC_DONE_TIMEOUT_MILLISEC = 500

	HADOOP_CONF_HDFS_SITE_FILENAME = "hdfs-site.xml"

	HADOOP_CONF_CORE_SITE_FILENAME = "core-site.xml"

	JINDO_MASTERNUM_DEFAULT = 1
	JINDO_HA_MASTERNUM      = 3

	DEFAULT_MASTER_RPC_PORT = 8101
	DEFAULT_WORKER_RPC_PORT = 6101
	DEFAULT_RAFT_RPC_PORT   = 8103

	POD_ROLE_TYPE = "role"

	WOKRER_POD_ROLE = "jindo-worker"

	runtimeFSType = "jindofs"
)
