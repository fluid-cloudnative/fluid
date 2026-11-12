/*
Copyright 2026 The Fluid Authors.

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

package jindo

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	OSSCredentialsProvider = "com.aliyun.jindodata.oss.auth.CustomCredentialsProvider"
	SecretProviderFormat   = "JSON"
	SecretMountPath        = "/token"
	AccessKeyIDSuffix      = ".accessKeyId"
	AccessKeySecretSuffix  = ".accessKeySecret"
	EndpointSuffix         = ".endpoint"
	DLSStorageEnableSuffix = ".data.lake.storage.enable"

	PlaintextCredentialWarning = "WARNING: OSS credentials will be stored in plaintext in ConfigMap"
)

func BuildBucketSecretURI(bucketName string) string {
	return fmt.Sprintf("secrets://%s/%s/", SecretMountPath, bucketName)
}

func ValidateSecretKeyRef(secretKeyRef datav1alpha1.SecretKeySelector, optionName, mountPoint string) error {
	if secretKeyRef.Name == "" || secretKeyRef.Key == "" {
		return fmt.Errorf("encryptOption %s for mount %s must reference both secret name and key", optionName, mountPoint)
	}

	return nil
}

func GetSecretDataValue(secret *corev1.Secret, secretName, secretKey string) (string, error) {
	value, ok := secret.Data[secretKey]
	if !ok {
		return "", fmt.Errorf("secret %s does not contain key %s", secretName, secretKey)
	}

	return string(value), nil
}

func AppendSecretProjection(projections []corev1.SecretProjection, secretName, secretKey, itemPath string) ([]corev1.SecretProjection, error) {
	for i, projection := range projections {
		for _, item := range projection.Items {
			if item.Path != itemPath {
				continue
			}
			if projection.Name == secretName && item.Key == secretKey {
				return projections, nil
			}
			return nil, fmt.Errorf("conflicting secret projection for %s", itemPath)
		}
		if projection.Name == secretName {
			projections[i].Items = append(projections[i].Items, corev1.KeyToPath{
				Key:  secretKey,
				Path: itemPath,
			})
			return projections, nil
		}
	}

	return append(projections, corev1.SecretProjection{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: secretName,
		},
		Items: []corev1.KeyToPath{{
			Key:  secretKey,
			Path: itemPath,
		}},
	}), nil
}

func SetBucketSecretProviderProperties(properties map[string]string, prefix, bucketName, secretURI string) {
	properties[fmt.Sprintf("%s.oss.bucket.%s.credentials.provider", prefix, bucketName)] = OSSCredentialsProvider
	properties[fmt.Sprintf("%s.oss.bucket.%s.provider.endpoint", prefix, bucketName)] = secretURI
	properties[fmt.Sprintf("%s.oss.bucket.%s.provider.format", prefix, bucketName)] = SecretProviderFormat
}

func BuildBucketPropertyKey(prefix, bucketName, suffix string) string {
	return prefix + bucketName + suffix
}

func ExtractSingleOSSEndpoint(properties map[string]string, prefix string) (string, bool) {
	var endpoint string
	for key, value := range properties {
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, EndpointSuffix) {
			continue
		}
		if endpoint == "" {
			endpoint = value
			continue
		}
		if endpoint != value {
			return "", false
		}
	}
	return endpoint, endpoint != ""
}

func ParseOSSMountBucket(mountPoint string) (string, error) {
	bucket, _, _ := strings.Cut(strings.TrimPrefix(mountPoint, "oss://"), "/")
	if bucket == "" {
		return "", fmt.Errorf("incorrect oss mountPoint with %v, please check your path is dir or file ", mountPoint)
	}

	return bucket, nil
}
