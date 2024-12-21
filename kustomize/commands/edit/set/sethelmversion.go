package set

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setHelmVersionOptions struct {
	chartMap map[string]string
}

func newCmdSetHelmVersion(fSys filesys.FileSystem) *cobra.Command {
	var o setHelmVersionOptions

	cmd := &cobra.Command{
		Use:   "helmversion",
		Short: `Sets helm chart versions in the kustomization file`,
		Example: `
The command
  set helmversion my-chart=1.2.3 my-other-chart=4.5.6

will edit the version of the helm charts in the kustomization file to the specified versions:

helmCharts:
- name: my-chart
  version: 1.2.3
  repo: oci://myrepo
- name: my-other-chart
  version: 4.5.6
  repo: oci://myrepo
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.validate(args)
			if err != nil {
				return err
			}
			return o.setHelmVersion(fSys)
		},
	}
	return cmd
}

func (o *setHelmVersionOptions) validate(args []string) error {
	if len(args) == 0 {
		return errors.New("no helm chart version specified")
	}

	o.chartMap = make(map[string]string)

	for _, arg := range args {
		// Expecting arg of form "chartName=version"
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid argument '%s', must be chartName=version", arg)
		}
		chartName := parts[0]
		version := parts[1]

		if chartName == "" || version == "" {
			return fmt.Errorf("invalid argument '%s', chartName and version must not be empty", arg)
		}

		o.chartMap[chartName] = version
	}
	return nil
}

func (o *setHelmVersionOptions) setHelmVersion(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}

	for i, hc := range m.HelmCharts {
		if newVersion, found := o.chartMap[hc.Name]; found {
			// Update the version in place
			hc.Version = newVersion
			m.HelmCharts[i] = hc
		}
	}

	return mf.Write(m)
}
