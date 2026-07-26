package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/numaproj/helm-charts/upgrade/internal"
	"github.com/numaproj/helm-charts/upgrade/internal/mirror"
)

const usage = `upgrade — helm-charts release helper for numaflow

Usage:
  upgrade [upgrade-charts]          Run the existing CRD/RBAC/SA/Chart.yaml/values.yaml refresh
                                    (default subcommand; honours NUMAFLOW_VERSION env var)
  upgrade sync [flags]              Mirror the 15 chart-only-templated files (configmaps,
                                    deployments, secrets, services) from upstream numaflow

sync flags:
  --to-version=vX.Y.Z       New numaflow version (defaults to $NUMAFLOW_VERSION)
  --from-version=vX.Y.Z     Previous numaflow version (defaults to Chart.yaml appVersion)
  --apply                   Write clean merges back to the chart files (default: dry-run)
  --only=a.yaml,b.yaml      Restrict to the listed file basenames

Environment:
  NUMAFLOW_VERSION          New numaflow version; equivalent to --to-version
`

func main() {
	subcommand := "upgrade-charts"
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		subcommand = args[0]
		args = args[1:]
	}

	switch subcommand {
	case "upgrade-charts":
		runUpgradeCharts()
	case "sync":
		runSync(args)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n\n%s", subcommand, usage)
		os.Exit(2)
	}
}

func runUpgradeCharts() {
	numaflowVersion := os.Getenv("NUMAFLOW_VERSION")
	if numaflowVersion == "" {
		log.Fatalln("Numaflow version is required (set NUMAFLOW_VERSION)")
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

func runSync(args []string) {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	toVersion := fs.String("to-version", os.Getenv("NUMAFLOW_VERSION"), "new numaflow version (default: $NUMAFLOW_VERSION)")
	fromVersion := fs.String("from-version", "", "previous numaflow version (default: Chart.yaml appVersion)")
	apply := fs.Bool("apply", false, "write clean merges back to chart files (default: dry-run)")
	only := fs.String("only", "", "comma-separated basenames to restrict the sync to")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	opts := mirror.SyncOptions{
		FromVersion: *fromVersion,
		ToVersion:   *toVersion,
		Apply:       *apply,
	}
	if *only != "" {
		opts.Only = strings.Split(*only, ",")
	}

	report, err := mirror.Run(opts)
	if err != nil {
		log.Fatalln("sync failed:", err)
	}
	if report.HasFailures() {
		os.Exit(1)
	}
}
