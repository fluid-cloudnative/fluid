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
	"fmt"
	"os"

	"github.com/fluid-cloudnative/fluid"
	"github.com/fluid-cloudnative/fluid/pkg/csi/fuse"
	"github.com/spf13/cobra"
)

var (
	endpoint string
	nodeID   string
	short    bool
)

var cmd = &cobra.Command{
	Use:   "csi",
	Short: "CSI based fluid driver for Fuse",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start fluid driver on node",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version information",
	Long: "print version information",
	Run: func(cmd *cobra.Command, args []string) {
		v := fluid.GetVersion()

		if short {
			fmt.Printf("version: %s\n", v)
			return
		}
		fmt.Printf("  BuildDate: %s\n", v.BuildDate)
		fmt.Printf("  GitCommit: %s\n", v.GitCommit)
		fmt.Printf("  GitTreeState: %s\n", v.GitTreeState)
		if v.GitTag != "" {
			fmt.Printf("  GitTag: %s\n", v.GitTag)
		}
		fmt.Printf("  GoVersion: %s\n", v.GoVersion)
		fmt.Printf("  Compiler: %s\n", v.Compiler)
		fmt.Printf("  Platform: %s\n", v.Platform)
	},
}

func init() {
	startCmd.Flags().StringVarP(&nodeID, "nodeid","","", "node id")
	if err := startCmd.MarkFlagRequired("nodeid"); err != nil {
		errorAndExit(err)
	}

	startCmd.Flags().StringVarP(&nodeID, "endpoint","","", "CSI endpoint")
	if err := startCmd.MarkFlagRequired("endpoint"); err != nil {
		errorAndExit(err)
	}

	versionCmd.Flags().BoolVar(&short, "short", false, "print just the short version info")

	cmd.AddCommand(startCmd)
	cmd.AddCommand(versionCmd)
}


func main() {
	if err := cmd.Execute(); err != nil {
		errorAndExit(err)
	}

	os.Exit(0)
}

func handle() {
	// startReaper()

	d := csi.NewDriver(nodeID, endpoint)
	d.Run()
}

func errorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s", err.Error())
	os.Exit(1)
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
