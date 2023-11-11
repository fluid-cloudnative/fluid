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

package base

import "testing"

func TestSafeClose(t *testing.T) {
	var nilCh chan MetadataSyncResult = nil

	openCh := make(chan MetadataSyncResult)

	closedCh := make(chan MetadataSyncResult)
	close(closedCh)

	tests := []struct {
		name       string
		ch         chan MetadataSyncResult
		wantClosed bool
	}{
		{
			name:       "close_open_channel",
			ch:         openCh,
			wantClosed: false,
		},
		{
			name:       "close_nil_channel",
			ch:         nilCh,
			wantClosed: false,
		},
		{
			name:       "close_closed_channel",
			ch:         closedCh,
			wantClosed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotClosed := SafeClose(tt.ch); gotClosed != tt.wantClosed {
				t.Errorf("SafeClose() = %v, want %v", gotClosed, tt.wantClosed)
			}
		})
	}
}

func TestSafeSend(t *testing.T) {
	var nilCh chan MetadataSyncResult = nil

	openCh := make(chan MetadataSyncResult)
	go func() {
		<-openCh
	}()

	closedCh := make(chan MetadataSyncResult)
	close(closedCh)

	type args struct {
		ch     chan MetadataSyncResult
		result MetadataSyncResult
	}
	tests := []struct {
		name       string
		args       args
		wantClosed bool
	}{
		{
			name: "send_to_open_channel",
			args: args{
				ch:     openCh,
				result: MetadataSyncResult{},
			},
			wantClosed: false,
		},
		{
			name: "send_to_nil_channel",
			args: args{
				ch:     nilCh,
				result: MetadataSyncResult{},
			},
			wantClosed: false,
		},
		{
			name: "send_to_closed_channel",
			args: args{
				ch:     closedCh,
				result: MetadataSyncResult{},
			},
			wantClosed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotClosed := SafeSend(tt.args.ch, tt.args.result); gotClosed != tt.wantClosed {
				t.Errorf("SafeSend() = %v, want %v", gotClosed, tt.wantClosed)
			}
		})
	}
}
