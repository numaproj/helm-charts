package main

import (
	"fmt"
	"log"
	"os"

	"github.com/numaproj/helm-charts/upgrade/internal"
)

func main() {
	numaflowVersion := os.Args[1]
	if numaflowVersion == "" {
		log.Fatalln("Numaflow version is required as a command line argument")
	}

	fmt.Println("################### Updating latest CRDs ###################")
	internal.UpdateCRDFiles(numaflowVersion)
	fmt.Println("\n################### Updating latest data for RBAC ###################")
	internal.UpdateRBACFiles(numaflowVersion)
	fmt.Println("\n################### Updating latest data for Service Account ###################")
	internal.UpdateServiceAccount(numaflowVersion)
}
