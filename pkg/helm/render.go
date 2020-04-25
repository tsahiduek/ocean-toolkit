package helm

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

type Options struct {
	Chart      string
	Name       string
	Namespace  string
	Values     []string
	ValueFiles []string
	Debug      bool
}

// Render renders templates of a provided Helm chart.
func Render(options *Options) ([]byte, error) {
	client := action.NewInstall(new(action.Configuration))
	client.DryRun = true
	client.Replace = true // skip the name check
	client.ClientOnly = true
	client.APIVersions = []string{}
	client.IncludeCRDs = true
	client.DisableHooks = true
	client.Devel = true
	client.Namespace = options.Namespace

	valueOpts := &values.Options{
		Values:     options.Values,
		ValueFiles: options.ValueFiles,
	}

	rel, err := runInstall([]string{options.Name, options.Chart}, client, valueOpts)
	if err != nil && !options.Debug {
		if rel != nil {
			return nil, fmt.Errorf("%w: use --debug flag to render out invalid yaml", err)
		}
		return nil, err
	}

	// We ignore a potential error here because, when the --debug flag was
	// specified, we always want to print the YAML, even if it is not valid.
	// The error is still returned afterwards.
	if rel != nil {
		var manifests bytes.Buffer
		manifests.WriteString(strings.TrimSpace(rel.Manifest))
		return manifests.Bytes(), nil
	}

	return nil, err
}

func runInstall(args []string, client *action.Install,
	valueOpts *values.Options) (*release.Release, error) {

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	releaseName, chartName, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = releaseName

	settings := cli.New()
	cp, err := client.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	providers := getter.Providers{getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	}}

	values, err := valueOpts.MergeValues(providers)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if err := isChartInstallable(chart); err != nil {
		return nil, err
	}

	if chart.Metadata.Deprecated {
		fmt.Fprintf(os.Stderr, "WARNING: chart is deprecated %s", chartName)
	}

	if deps := chart.Metadata.Dependencies; deps != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chart, deps); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:        os.Stdout,
					ChartPath:  cp,
					Keyring:    client.ChartPathOptions.Keyring,
					SkipUpdate: false,
					Debug:      true,
					Getters:    providers,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chart, err = loader.Load(cp); err != nil {
					return nil, fmt.Errorf("failed reloading chart after repo update: %w", err)
				}
			} else {
				return nil, err
			}
		}
	}

	return client.Run(chart, values)
}

func isChartInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}
