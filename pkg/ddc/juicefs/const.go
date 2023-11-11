/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
