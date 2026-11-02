// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"os"
	"regexp"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// TestAlertmanagerCORSOriginPattern verifies that the generated Alertmanager
// CRD validates spec.web.corsOrigin with the expected pattern.
func TestAlertmanagerCORSOriginPattern(t *testing.T) {
	b, err := os.ReadFile("../../../../example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.UnmarshalStrict(b, &crd); err != nil {
		t.Fatal(err)
	}

	var pattern string
	for _, v := range crd.Spec.Versions {
		web, ok := v.Schema.OpenAPIV3Schema.Properties["spec"].Properties["web"]
		if !ok {
			continue
		}
		corsOrigin, ok := web.Properties["corsOrigin"]
		if !ok {
			t.Fatalf("version %s: spec.web.corsOrigin not found in CRD schema", v.Name)
		}
		if corsOrigin.Pattern == "" {
			t.Fatalf("version %s: spec.web.corsOrigin has no pattern", v.Name)
		}
		if pattern == "" {
			pattern = corsOrigin.Pattern
		} else if pattern != corsOrigin.Pattern {
			t.Fatalf("version %s: inconsistent corsOrigin pattern %q != %q", v.Name, corsOrigin.Pattern, pattern)
		}
	}
	if pattern == "" {
		t.Fatal("spec.web not found in any CRD version")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("pattern %q doesn't compile: %s", pattern, err)
	}

	for _, good := range []string{
		`https://example\.com`,
		`https?://(foo|bar)\.example\.com`,
		`https://.*\.example\.com`,
	} {
		if !re.MatchString(good) {
			t.Errorf("expected pattern %q to match %q", pattern, good)
		}
	}

	for _, bad := range []string{
		"",
		"https://example .com",
		" https://example\\.com",
		"https://example\\.com ",
		"https://example\n.com",
	} {
		if re.MatchString(bad) {
			t.Errorf("expected pattern %q not to match %q", pattern, bad)
		}
	}
}
