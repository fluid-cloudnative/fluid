/*
<<<<<<< HEAD
=======
Copyright 2021 The Fluid Authors.
>>>>>>> master

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
<<<<<<< HEAD
=======

>>>>>>> master
package app

import "github.com/spf13/cobra"

<<<<<<< HEAD
func NewFluidCSICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fluid-csi",
		Short: "CSI based fluid driver for Fuse",
	}

	cmd.AddCommand(versionCmd, csiCmd)

=======
func NewCSICommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "fluid-csi",
		Short: "CSI based fluid driver for Fuse",
	}
	cmd.AddCommand(startCmd)
	cmd.AddCommand(versionCmd)
>>>>>>> master
	return cmd
}
