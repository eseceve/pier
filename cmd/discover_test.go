package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/comparaonline/pier/internal/discover"
)

// fakeCmdRunner returns canned kubectl output keyed by joined args.
type fakeCmdRunner struct{ out map[string]string }

func (f fakeCmdRunner) Output(args []string) ([]byte, error) {
	return []byte(f.out[strings.Join(args, " ")]), nil
}

func withDiscover(t *testing.T, r discover.CmdRunner) {
	t.Helper()
	prev := cmdRunner
	cmdRunner = r
	discoverJSON, discoverContext, discoverNamespace = false, "", ""
	t.Cleanup(func() { cmdRunner = prev })
}

func TestDiscoverJSONListsContexts(t *testing.T) {
	withDiscover(t, fakeCmdRunner{out: map[string]string{
		"config get-contexts -o name": "staging-ctx\nprod-ctx\n",
	}})

	rootCmd.SetArgs([]string{"discover", "--json"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, `"contexts"`) || !strings.Contains(s, "staging-ctx") {
		t.Fatalf("got %q", s)
	}
}

func TestDiscoverJSONListsResourcesFiltered(t *testing.T) {
	withDiscover(t, fakeCmdRunner{out: map[string]string{
		"--context c --namespace n get deployments -o name": "deployment.apps/api\n",
		"--context c --namespace n get configmaps -o name":  "configmap/api-config\n",
		"--context c --namespace n get secrets -o name":     "secret/api-local\nsecret/sh.helm.release.v1.api.v1\n",
	}})

	rootCmd.SetArgs([]string{"discover", "--json", "--context", "c", "--namespace", "n"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "api-local") || !strings.Contains(s, "api-config") {
		t.Fatalf("missing resources: %q", s)
	}
	if strings.Contains(s, "helm") {
		t.Fatalf("helm secret should be filtered: %q", s)
	}
}
