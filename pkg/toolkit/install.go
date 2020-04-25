package toolkit

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	oceanv1 "github.com/spotinst/ocean-toolkit/api/ocean/v1"
	"github.com/spotinst/ocean-toolkit/pkg/helm"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// NamespaceOceanSystem is the name of the namespace where we place Ocean components.
const NamespaceOceanSystem string = "ocean-system"

// Default returns the default Ocean Toolkit spec.
func Default() *oceanv1.OceanToolkit {
	return &oceanv1.OceanToolkit{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ocean.spot.io.Toolkit",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ocean-toolkit",
			Namespace: NamespaceOceanSystem,
		},
		Spec: oceanv1.OceanToolkitSpec{
			Components: map[string]*oceanv1.OceanToolkitComponent{
				"controller": {
					Enabled: true,
				},
				"operator": {
					Enabled: true,
				},
				"metrics-server": {
					Enabled: true,
				},
			},
			Values: map[string]interface{}{
				"spot": map[string]interface{}{
					"token":   os.Getenv(credentials.EnvCredentialsVarToken),
					"account": os.Getenv(credentials.EnvCredentialsVarAccount),
				},
			},
		},
	}
}

// Options contains all possible options for Ocean Toolkit installation.
type Options struct {
	Writer     io.Writer
	Values     []string
	ValueFiles []string
	Namespace  string
	Chart      string
}

// Install install the Ocean Toolkit.
func Install(ctx context.Context, toolkit *oceanv1.OceanToolkit, options *Options) error {
	renderValues := convertOceanToolkitSpecToValues(toolkit.Spec)

	b, err := yaml.Marshal(renderValues)
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile(os.TempDir(), "ocean-toolkit")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	f.Write(b)
	f.Close()

	chart := options.Chart
	if chart == "" {
		chart = chartPath("toolkit", "https://spotinst.github.io/ocean-charts/releases/stable", "0.0.1")
	}

	renderOpts := &helm.Options{
		Chart:      chart,
		Name:       "toolkit",
		Debug:      true,
		Values:     options.Values,
		ValueFiles: []string{f.Name()},
		Namespace:  options.Namespace,
	}

	b, err = helm.Render(renderOpts)
	if err != nil {
		return err
	}

	if options.Writer != nil {
		options.Writer.Write(b)
	}

	return nil
}

func convertOceanToolkitSpecToValues(in oceanv1.OceanToolkitSpec) map[string]interface{} {
	out := make(map[string]interface{})

	// Copy global values.
	if len(in.Values) > 0 {
		out["global"] = in.Values
	}

	for name, comp := range in.Components {
		// Deploy?
		values := map[string]interface{}{
			"enabled": comp.Enabled,
		}

		// Copy local values.
		if len(comp.Values) > 0 {
			for k, v := range comp.Values {
				values[k] = v
			}
		}

		// Add the component.
		out[name] = values
	}

	return out
}

func chartPath(chart, repo, version string) string {
	return fmt.Sprintf("%s/%s-%s.tgz", repo, chart, version)
}
