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

import "testing"

func TestSplitSchemaAddr(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name         string
		args         args
		wantProtocol string
		wantAddr     string
	}{
		{
			name: "Test for unix protocol",
			args: args{
				addr: "unix:///foo/bar",
			},
			wantProtocol: "unix",
			wantAddr:     "/foo/bar",
		},
		{
			name: "Test for tcp protocol",
			args: args{
				addr: "tcp://127.0.0.1:8088",
			},
			wantProtocol: "tcp",
			wantAddr:     "127.0.0.1:8088",
		},
		{
			name: "Test for default protocol",
			args: args{
				addr: "127.0.0.1:3456",
			},
			wantProtocol: "tcp",
			wantAddr:     "127.0.0.1:3456",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProtocol, gotAddr := SplitSchemaAddr(tt.args.addr)
			if gotProtocol != tt.wantProtocol {
				t.Errorf("SplitSchemaAddr() gotProtocol = %v, want %v", gotProtocol, tt.wantProtocol)
			}
			if gotAddr != tt.wantAddr {
				t.Errorf("SplitSchemaAddr() gotAddr = %v, want %v", gotAddr, tt.wantAddr)
			}
		})
	}
}
