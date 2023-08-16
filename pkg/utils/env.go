/*
Copyright 2022 The Fluid Authors.

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

package utils

import (
	"os"
	"strconv"
	"time"
)

func GetDurationValueFromEnv(key string, defaultValue time.Duration) (value time.Duration) {
	var err error
	value = defaultValue

	str, ok := os.LookupEnv(key)
	// if not set, return the default value
	if !ok {
		return
	}

	value, err = time.ParseDuration(str)
	if err != nil {
		value = defaultValue
	}

	return
}

func GetBoolValueFromEnv(key string, defaultValue bool) (value bool) {
	value = defaultValue
	var err error

	str, ok := os.LookupEnv(key)
	// if not set, return the default value
	if !ok {
		return
	}

	value, err = strconv.ParseBool(str)
	if err != nil {
		value = defaultValue
	}
	return
}

func GetIntValueFromEnv(key string) (value int, found bool) {

	str, found := os.LookupEnv(key)
	// if not set, return the default value
	if !found {
		return
	}

	value, err := strconv.Atoi(str)
	if err != nil {
		found = false
	}
	return
}

func GetStringValueFromEnv(key string, defaultValue string) (value string) {
	if res, found := os.LookupEnv(key); found {
		return res
	}

	return defaultValue
}
