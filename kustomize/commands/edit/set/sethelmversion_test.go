package set

import (
	"fmt"
	"strings"
	"testing"

	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestSetHelmVersion(t *testing.T) {
	type given struct {
		args             []string
		infileHelmCharts []string
	}
	type expected struct {
		fileOutput []string
		err        error
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			description: "error no args",
			given: given{
				args: []string{},
			},
			expected: expected{
				err: fmt.Errorf("no helm chart version specified"),
			},
		},
		{
			description: "invalid format no equals",
			given: given{
				args: []string{"mychart1.2.3"},
			},
			expected: expected{
				err: fmt.Errorf("invalid argument 'mychart1.2.3', must be chartName=version"),
			},
		},
		{
			description: "empty chart name",
			given: given{
				args: []string{"=1.2.3"},
			},
			expected: expected{
				err: fmt.Errorf("invalid argument '=1.2.3', chartName and version must not be empty"),
			},
		},
		{
			description: "empty version",
			given: given{
				args: []string{"mychart="},
			},
			expected: expected{
				err: fmt.Errorf("invalid argument 'mychart=', chartName and version must not be empty"),
			},
		},
		{
			description: "valid single argument and one chart",
			given: given{
				args: []string{"mychart=1.2.3"},
				infileHelmCharts: []string{
					"helmCharts:",
					"- name: mychart",
					"  version: old-version",
					"  repo: oci://myrepo",
				},
			},
			expected: expected{
				fileOutput: []string{
					"helmCharts:",
					"- name: mychart",
					"  repo: oci://myrepo",
					"  version: 1.2.3",
				},
			},
		},
		{
			description: "valid multiple arguments and multiple distinct charts",
			given: given{
				args: []string{"chartA=2.0.0", "chartB=3.0.0"},
				infileHelmCharts: []string{
					"helmCharts:",
					"- name: chartA",
					"  version: oldA",
					"- name: chartB",
					"  version: oldB",
					"- name: chartC",
					"  version: unchangedC",
				},
			},
			expected: expected{
				fileOutput: []string{
					"helmCharts:",
					"- name: chartA",
					"  version: 2.0.0",
					"- name: chartB",
					"  version: 3.0.0",
					"- name: chartC",
					"  version: unchangedC",
				},
			},
		},
		{
			description: "multiple charts with same name updated",
			given: given{
				args: []string{"redis=10.0.0"},
				infileHelmCharts: []string{
					"helmCharts:",
					"- name: redis",
					"  version: old-version",
					"- name: redis",
					"  version: another-old-version",
					"- name: filing",
					"  version: stay-the-same",
				},
			},
			expected: expected{
				fileOutput: []string{
					"helmCharts:",
					"- name: redis",
					"  version: 10.0.0",
					"- name: redis",
					"  version: 10.0.0",
					"- name: filing",
					"  version: stay-the-same",
				},
			},
		},
		{
			description: "no helmCharts field in file",
			given: given{
				args: []string{"mychart=2.2.2"},
				// no helmCharts in input
			},
			expected: expected{
				// no helmCharts added or modified; no error
				fileOutput: []string{},
			},
		},
		{
			description: "chart does not exist",
			given: given{
				args: []string{"nonexistent=9.9.9"},
				infileHelmCharts: []string{
					"helmCharts:",
					"- name: chartA",
					"  version: oldA",
				},
			},
			expected: expected{
				// no changes since chart doesn't match
				fileOutput: []string{
					"helmCharts:",
					"- name: chartA",
					"  version: oldA",
				},
			},
		},
		{
			description: "update chart in complex scenario",
			given: given{
				args: []string{"gui=5.0.0", "hermes=2.0.0"},
				infileHelmCharts: []string{
					"helmCharts:",
					"- name: miccertificate",
					"  version: 1.0.0",
					"- name: gui",
					"  version: 1.0.0",
					"- name: redis",
					"  version: old",
					"- name: hermes",
					"  version: old-hermes",
				},
			},
			expected: expected{
				fileOutput: []string{
					"helmCharts:",
					"- name: miccertificate",
					"  version: 1.0.0",
					"- name: gui",
					"  version: 5.0.0",
					"- name: redis",
					"  version: old",
					"- name: hermes",
					"  version: 2.0.0",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s%v", tc.description, tc.given.args), func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			cmd := newCmdSetHelmVersion(fSys)

			if len(tc.given.infileHelmCharts) > 0 {
				// write file with infileHelmCharts
				testutils_test.WriteTestKustomizationWith(
					fSys,
					[]byte(strings.Join(tc.given.infileHelmCharts, "\n")))
			} else {
				// Write empty or base kustomization if no helmCharts provided
				testutils_test.WriteTestKustomization(fSys)
			}

			// Run the command
			err := cmd.RunE(cmd, tc.given.args)

			// Check error expectations
			if tc.expected.err == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
				t.FailNow()
			}
			if tc.expected.err != nil && (err == nil || err.Error() != tc.expected.err.Error()) {
				t.Errorf("Expected error: %v, got: %v", tc.expected.err, err)
				t.FailNow()
			}

			// If there's expected file output, verify it
			if len(tc.expected.fileOutput) > 0 {
				content, err := testutils_test.ReadTestKustomization(fSys)
				if err != nil {
					t.Errorf("unexpected read error: %v", err)
					t.FailNow()
				}
				expectedStr := strings.Join(tc.expected.fileOutput, "\n")
				if !strings.Contains(string(content), expectedStr) {
					t.Errorf("unexpected helm charts in kustomization file. \nActual:\n%s\nExpected:\n%s", content, expectedStr)
				}
			}
		})
	}
}
