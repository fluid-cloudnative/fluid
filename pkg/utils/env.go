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
