/*

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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/csi/fuse"
	"github.com/spf13/cobra"
)

var (
	endpoint string
	nodeID   string
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {

	flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   "fluid",
		Short: "CSI based fluid driver for Fuse",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&nodeID, "nodeid", "", "node id")
	cmd.MarkPersistentFlagRequired("nodeid")

	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "CSI endpoint")
	cmd.MarkPersistentFlagRequired("endpoint")

	cmd.ParseFlags(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func handle() {
	// startReaper()

	d := csi.NewDriver(nodeID, endpoint)
	d.Run()
}

// /*
// Based on https://github.com/openshift/origin/blob/master/pkg/util/proc/reaper.go

// */
// func startReaper() {
// 	glog.V(4).Infof("Launching reaper")
// 	go func() {
// 		sigs := make(chan os.Signal, 1)
// 		signal.Notify(sigs, syscall.SIGCHLD)
// 		for {
// 			// Wait for a child to terminate
// 			sig := <-sigs
// 			glog.V(4).Infof("Signal received: %v", sig)
// 			for {
// 				// Reap processes
// 				cpid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
// 				if cpid < 1 {
// 					break
// 				}

// 				glog.V(4).Infof("Reaped process with pid %d", cpid)
// 			}
// 		}
// 	}()
// }
