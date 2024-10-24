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
