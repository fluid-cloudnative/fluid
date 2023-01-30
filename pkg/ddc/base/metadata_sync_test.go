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
