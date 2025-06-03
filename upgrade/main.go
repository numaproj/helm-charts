package main

import (
	"log"
	"os"

	"github.com/numaproj/helm-charts/upgrade/internal"
)

func main() {
	numaflowVersion := os.Getenv("NUMAFLOW_VERSION")
	if numaflowVersion == "" {
		log.Fatalln("Numaflow version is required as a command line argument")
	}

	log.Println("Checking version existence in Numaflow repo...")
	if exists, err := internal.IsVersionExists(numaflowVersion); !exists && err != nil {
		log.Fatalln("Version check failed:", err)
	}

	log.Println("################### Updating Chart.yaml ###################")
	internal.UpdateChartFile(numaflowVersion)

	log.Println("################### Updating Values.yaml ###################")
	internal.UpdateValuesFile(numaflowVersion)

	log.Println("\n################### Updating latest CRDs ###################")
	internal.UpdateCRDFiles(numaflowVersion)

	log.Println("\n################### Updating latest data for RBAC ###################")
	internal.UpdateRBACFiles(numaflowVersion)

	log.Println("\n################### Updating latest data for Service Account ###################")
	internal.UpdateServiceAccount(numaflowVersion)
}
