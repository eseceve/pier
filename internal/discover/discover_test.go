package discover

import (
	"reflect"
	"strings"
	"testing"
)

// fakeRunner returns canned output keyed by the joined kubectl args.
type fakeRunner struct{ out map[string]string }

func (f fakeRunner) Output(args []string) ([]byte, error) {
	return []byte(f.out[strings.Join(args, " ")]), nil
}

func TestContexts(t *testing.T) {
	r := fakeRunner{out: map[string]string{
		"config get-contexts -o name": "staging-ctx\nprod-ctx\n",
	}}
	got, err := Contexts(r)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"staging-ctx", "prod-ctx"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDeployments(t *testing.T) {
	r := fakeRunner{out: map[string]string{
		"--context c --namespace n get deployments -o name": "deployment.apps/api\ndeployment.apps/web\n",
	}}
	got, err := Deployments(r, "c", "n")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api", "web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSecretsAreFiltered(t *testing.T) {
	r := fakeRunner{out: map[string]string{
		"--context c --namespace n get secrets -o name": "secret/api-local\nsecret/sh.helm.release.v1.api.v1\n",
	}}
	got, err := Secrets(r, "c", "n")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api-local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
