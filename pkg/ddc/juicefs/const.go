/*
Copyright 2021 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package juicefs

const (
	BlockCacheBytesOfEnterprise     = "blockcache.bytes"
	BlockCacheHitsOfEnterprise      = "blockcache.hits"
	BlockCacheMissOfEnterprise      = "blockcache.miss"
	BlockCacheHitBytesOfEnterprise  = "blockcache.hitBytes"
	BlockCacheMissBytesOfEnterprise = "blockcache.missBytes"

	BlockCacheBytesOfCommunity     = "juicefs_blockcache_bytes"
	BlockCacheHitsOfCommunity      = "juicefs_blockcache_hits"
	BlockCacheMissOfCommunity      = "juicefs_blockcache_miss"
	BlockCacheHitBytesOfCommunity  = "juicefs_blockcache_hit_bytes"
	BlockCacheMissBytesOfCommunity = "juicefs_blockcache_miss_bytes"

	workerPodRole      = "juicefs-worker"
	EnterpriseEdition  = "enterprise"
	CommunityEdition   = "community"
	DefaultMetricsPort = 9567

	MetadataSyncNotDoneMsg               = "[Calculating]"
	CheckMetadataSyncDoneTimeoutMillisec = 500

	DefaultCacheDir = "/var/jfsCache"

	JuiceStorage   = "storage"
	JuiceBucket    = "bucket"
	JuiceMetaUrl   = "metaurl"
	JuiceAccessKey = "access-key"
	JuiceSecretKey = "secret-key"
	JuiceToken     = "token"

	MountPath                 = "mountpath"
	Edition                   = "edition"
	MetaurlSecret             = "metaurlSecret"
	MetaurlSecretKey          = "metaurlSecretKey"
	TokenSecret               = "tokenSecret"
	TokenSecretKey            = "tokenSecretKey"
	AccessKeySecret           = "accesskeySecret"
	AccessKeySecretKey        = "accesskeySecretKey"
	SecretKeySecret           = "secretkeySecret"
	SecretKeySecretKey        = "secretkeySecretKey"
	FormatCmd                 = "formatCmd"
	Name                      = "name"
	DefaultDataLoadTimeout    = "30m"
	DefaultDataMigrateTimeout = "30m"

	NativeVolumeMigratePath = "/mnt/fluid-native/"
)
