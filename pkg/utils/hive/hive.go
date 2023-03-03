package hive

import (
	"github.com/dazheng/gohive"
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateHiveClient create the hive client
func CreateHiveClient(host string) (conn *gohive.Connection, err error) {
	connect, err := gohive.Connect(host, gohive.DefaultOptions)
	return connect, err
}

// ChangeSchemaURLForRecover recover the table location
func ChangeSchemaURLForRecover(client client.Client, datatable data.DataTable, conn *gohive.Connection) error {
	return nil
}
