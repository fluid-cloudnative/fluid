/*
Copyright 2024 The Fluid Authors.

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

package operations

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestAlluxioFileUtils_DecommissionWorkers(t *testing.T) {
	a := &AlluxioFileUtils{log: fake.NullLogger()}

	t.Run("empty address list is a no-op", func(t *testing.T) {
		if err := a.DecommissionWorkers(nil); err != nil {
			t.Fatalf("want nil, got: %v", err)
		}
		if err := a.DecommissionWorkers([]string{}); err != nil {
			t.Fatalf("want nil, got: %v", err)
		}
	})

	t.Run("exec error is propagated", func(t *testing.T) {
		patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec,
			func(_ AlluxioFileUtils, _ []string, _ bool) (string, string, error) {
				return "", "", errors.New("exec failed")
			})
		defer patches.Reset()

		if err := a.DecommissionWorkers([]string{"192.168.1.1:29999"}); err == nil {
			t.Error("want error, got nil")
		}
	})

	t.Run("address is forwarded to the alluxio CLI", func(t *testing.T) {
		var capturedCmd []string
		patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec,
			func(_ AlluxioFileUtils, cmd []string, _ bool) (string, string, error) {
				capturedCmd = cmd
				return "", "", nil
			})
		defer patches.Reset()

		addr := "192.168.1.1:29999"
		if err := a.DecommissionWorkers([]string{addr}); err != nil {
			t.Fatalf("want nil, got: %v", err)
		}
		found := false
		for _, arg := range capturedCmd {
			if arg == addr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("address %q not found in command: %v", addr, capturedCmd)
		}
	})

	t.Run("multiple addresses are joined with commas", func(t *testing.T) {
		var capturedCmd []string
		patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec,
			func(_ AlluxioFileUtils, cmd []string, _ bool) (string, string, error) {
				capturedCmd = cmd
				return "", "", nil
			})
		defer patches.Reset()

		if err := a.DecommissionWorkers([]string{"10.0.0.1:29999", "10.0.0.2:29999"}); err != nil {
			t.Fatalf("want nil, got: %v", err)
		}
		found := false
		for _, arg := range capturedCmd {
			if arg == "10.0.0.1:29999,10.0.0.2:29999" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("joined addresses not found in command: %v", capturedCmd)
		}
	})
}

func TestAlluxioFileUtils_CountActiveWorkers(t *testing.T) {
	a := &AlluxioFileUtils{log: fake.NullLogger()}

	t.Run("exec error returns zero and the error", func(t *testing.T) {
		patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec,
			func(_ AlluxioFileUtils, _ []string, _ bool) (string, string, error) {
				return "", "", errors.New("exec failed")
			})
		defer patches.Reset()

		count, err := a.CountActiveWorkers()
		if err == nil {
			t.Error("want error, got nil")
		}
		if count != 0 {
			t.Errorf("want 0 on error, got %d", count)
		}
	})

	t.Run("two active workers", func(t *testing.T) {
		report := `Capacity information for all workers:
   Total Capacity: 4096.00MB
   Used Capacity: 443.89MB

Worker Name      Last Heartbeat   Storage       MEM
192.168.1.147    0                capacity      2048.00MB
                                 used          443.89MB (21%)
192.168.1.146    0                capacity      2048.00MB
                                 used          0B (0%)
`
		patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec,
			func(_ AlluxioFileUtils, _ []string, _ bool) (string, string, error) {
				return report, "", nil
			})
		defer patches.Reset()

		count, err := a.CountActiveWorkers()
		if err != nil {
			t.Fatalf("want nil, got: %v", err)
		}
		if count != 2 {
			t.Errorf("want 2, got %d", count)
		}
	})
}

func TestParseActiveWorkerCount(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "empty report",
			input:  "",
			expect: 0,
		},
		{
			name:   "no worker section header",
			input:  "Capacity information for all workers:\n   Total Capacity: 0B\n",
			expect: 0,
		},
		{
			name: "single worker",
			input: `Worker Name      Last Heartbeat   Storage       MEM
192.168.1.1      0                capacity      1024.00MB
                                 used          0B (0%)
`,
			expect: 1,
		},
		{
			name: "three workers",
			input: `Worker Name      Last Heartbeat   Storage       MEM
10.0.0.1         0                capacity      2048.00MB
                                 used          100MB (5%)
10.0.0.2         0                capacity      2048.00MB
                                 used          0B (0%)
10.0.0.3         0                capacity      2048.00MB
                                 used          500MB (25%)
`,
			expect: 3,
		},
		{
			name: "trailing blank lines are ignored",
			input: `Worker Name      Last Heartbeat   Storage       MEM
10.0.0.1         0                capacity      1024.00MB
                                 used          0B (0%)


`,
			expect: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseActiveWorkerCount(tc.input)
			if got != tc.expect {
				t.Errorf("want %d, got %d", tc.expect, got)
			}
		})
	}
}
