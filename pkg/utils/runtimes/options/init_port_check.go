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

package options

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	EnvPortCheckEnabled = "INIT_PORT_CHECK_ENABLED"
)

var initPortCheckEnabled = false

func setPortCheckOption() {
	if strVal, found := os.LookupEnv(EnvPortCheckEnabled); found {
		if boolVal, err := strconv.ParseBool(strVal); err != nil {
			panic(errors.Wrapf(err, "can't parse %s to bool", EnvPortCheckEnabled))
		} else {
			initPortCheckEnabled = boolVal
		}
	}
	// log.Printf("Using %s = %v\n", EnvPortCheckEnabled, initPortCheckEnabled)
	log.Info("setPortCheckOption", "EnvPortCheckEnabled",
		EnvPortCheckEnabled,
		"initPortCheckEnabled",
		initPortCheckEnabled)
}

func PortCheckEnabled() bool {
	return initPortCheckEnabled
}
