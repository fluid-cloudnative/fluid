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

package generator

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"
)

// ValidCACert think cert and key are valid if they meet the following requirements:
// - key and cert are valid pair
// - caCert is the root ca of cert
// - cert is for dnsName
// - cert won't expire before time
func ValidCACert(key, cert, caCert []byte, dnsName string, time time.Time) bool {
	if len(key) == 0 || len(cert) == 0 || len(caCert) == 0 {
		return false
	}
	// Verify key and cert are valid pair
	_, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return false
	}

	// Verify cert is valid for at least 1 year.
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return false
	}
	block, _ := pem.Decode(cert)
	if block == nil {
		return false
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	ops := x509.VerifyOptions{
		DNSName:     dnsName,
		Roots:       pool,
		CurrentTime: time,
	}
	_, err = c.Verify(ops)
	return err == nil
}
