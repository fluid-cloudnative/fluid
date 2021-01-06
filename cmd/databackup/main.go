package main

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"strings"
)

var(
	short                bool
	pvc                  bool    // store in pvc or local
	dataset              string
	namespace            string
	subPath              string
)

var cmd = &cobra.Command{
	Use:   "databackup",
	Short: "tool for databackup metadata",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start databackup tool in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fluid.PrintVersion(short)
	},
}

func init() {
	startCmd.Flags().StringVarP(&dataset, "dataset", "", "unknown", "The dataset been backup.")
	startCmd.Flags().StringVarP(&namespace, "namespace", "", "default", "the namespace of dataset.")
	startCmd.Flags().StringVarP(&subPath, "subpath", "", "", "the subPath of save.")
	startCmd.Flags().BoolVar(&pvc, "pvc", false, "store the gz file in pvc or not")
	versionCmd.Flags().BoolVar(&short, "short", false, "print just the short version info")
	cmd.AddCommand(startCmd)
	cmd.AddCommand(versionCmd)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
}

func handle() {
	// 1. exec backup command
	binary := "alluxio"
	args := []string{"fsadmin", "backup"}

	cmd := exec.Command(binary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err, "failed to execute", "args", strings.Join(args, " "))
	}

	// 2. get the new path
	newPath, err := getNewPath()
	if err != nil {
		log.Println(err, "path is fault, will not backup, please check!")
		os.Exit(1)
	}

	// 3. get the backup result by splitting str
	strs := strings.Split(string(out), "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, cdatabackup.BACKUP_RESULT_BACKUP_URI) {
			// 4. need to change metadata file path and filename
			binary = "mv"
			oldPathAndFileName := strings.TrimPrefix(str, cdatabackup.BACKUP_RESULT_BACKUP_URI)
			args = []string{oldPathAndFileName, newPath + getGzFilename()}
			cmd = exec.Command(binary, args...)
			_, err = cmd.CombinedOutput()
			if err != nil {
				log.Println(err, "failed to execute", "args", strings.Join(args, " "))
			}
		}
	}

	// 5. need to change metadata info file path and filename
	binary = "cp"
	oldPath := cdatabackup.BACPUP_PATH_POD + "/"
	MetadataInfoFilename := getMetadataInfoFilename()
	args = []string{oldPath + MetadataInfoFilename, newPath + MetadataInfoFilename}
	cmd = exec.Command(binary, args...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(err, "failed to execute", "args", strings.Join(args, " "))
	}
}

func getGzFilename() string {
	return "metadata-backup-" + dataset + "-" + namespace + ".gz"
}

func getMetadataInfoFilename() string {
	return dataset + "-" + namespace + ".yaml"
}

func getNewPath() (newPath string, err error) {
	// deside rootPath according to pvc or not
	rootPath := ""
	if pvc{
		rootPath = cdatabackup.PVC_PATH_POD + "/"

	} else{
		rootPath = cdatabackup.BACPUP_PATH_POD + "/"
	}

	// handle the subpath
	if strings.HasPrefix(subPath, "/"){
		subPath = strings.TrimPrefix(subPath, "/")
	}
	if !strings.HasSuffix(subPath, "/"){
		subPath = subPath + "/"
	}
	newPath = rootPath + subPath
	//check path exists or not
	_, err = os.Stat(newPath)
	if err != nil {
		binary := "mkdir"
		args := []string{"-p", newPath}
		cmd := exec.Command(binary, args...)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err, "failed to execute", "args", strings.Join(args, " "))
		}

	}
	//check path exists or not again
	_, err = os.Stat(newPath)
	return
}
