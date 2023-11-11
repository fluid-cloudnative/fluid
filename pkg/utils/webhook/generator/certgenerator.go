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

// Artifacts hosts a private key, its corresponding serving certificate and
// the CA certificate that signs the serving certificate.
type Artifacts struct {
	// PEM encoded private key
	Key []byte
	// PEM encoded serving certificate
	Cert []byte
	// PEM encoded CA private key
	CAKey []byte
	// PEM encoded CA certificate
	CACert []byte
	// Resource version of the certs
	ResourceVersion string
}

// CertGenerator is an interface to provision the serving certificate.
type CertGenerator interface {
	// Generate returns a Artifacts struct.
	Generate(CommonName string) (*Artifacts, error)
	// SetCA sets the PEM-encoded CA private key and CA cert for signing the generated serving cert.
	SetCA(caKey, caCert []byte)
}
