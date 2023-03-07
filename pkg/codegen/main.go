package main

import (
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
	cdicore "kubevirt.io/containerized-data-importer-api/pkg/apis/core"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func main() {
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/wrangler/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			cdicore.GroupName: {
				Types: []interface{}{
					cdiv1.DataVolume{},
					cdiv1.DataVolumeList{},
					cdiv1.CDIConfig{},
					cdiv1.CDIConfigList{},
					cdiv1.CDI{},
					cdiv1.CDIList{},
					cdiv1.StorageProfile{},
					cdiv1.StorageProfileList{},
					cdiv1.DataSource{},
					cdiv1.DataSourceList{},
					cdiv1.DataImportCron{},
					cdiv1.DataImportCronList{},
					cdiv1.ObjectTransfer{},
					cdiv1.ObjectTransferList{},
				},
				InformersPackage: "kubevirt.io/containerized-data-importer/pkg/client/informers",
				ClientSetPackage: "kubevirt.io/containerized-data-importer/pkg/client/clientset",
				ListersPackage:   "kubevirt.io/containerized-data-importer/pkg/client/listers",
			},
		},
	})
}
