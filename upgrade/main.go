package main

import (
	"fmt"
	"github.com/numaproj/helm-charts/upgrade/internal"
	"log"
	"os"
)

func main() {
	numaflowVersion := os.Args[1]
	if numaflowVersion == "" {
		log.Fatalln("Numaflow version is required as a command line argument")
	}

	fmt.Println("Checking version existence in Numaflow repo...")
	if exists, err := internal.IsVersionExists(numaflowVersion); !exists && err != nil {
		log.Fatalln("Version check failed:", err)
	}

	fmt.Println("################### Updating Chart.yaml ###################")
	internal.UpdateChartFile(numaflowVersion)
	fmt.Println("\n################### Updating latest CRDs ###################")
	internal.UpdateCRDFiles(numaflowVersion)
	fmt.Println("\n################### Updating latest data for RBAC ###################")
	internal.UpdateRBACFiles(numaflowVersion)
	fmt.Println("\n################### Updating latest data for Service Account ###################")
	internal.UpdateServiceAccount(numaflowVersion)
}
