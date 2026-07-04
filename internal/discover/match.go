// Package discover inspects a Kubernetes cluster via kubectl to help build
// pier's config.
package discover

import "strings"

// parseNames splits `kubectl ... -o name` output into bare names, stripping any
// "kind/" prefix and blank lines.
func parseNames(out []byte) []string {
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i := strings.LastIndex(line, "/"); i >= 0 {
			line = line[i+1:]
		}
		names = append(names, line)
	}
	return names
}

// filterNoise drops Kubernetes-managed names that never map to a pier service:
// helm release secrets, service-account tokens, and cluster CA material.
func filterNoise(names []string) []string {
	var out []string
	for _, n := range names {
		switch {
		case strings.HasPrefix(n, "sh.helm.release."):
		case strings.Contains(n, "-token-"):
		case n == "kube-root-ca.crt":
		case strings.HasPrefix(n, "istio-ca"):
		default:
			out = append(out, n)
		}
	}
	return out
}

// Match returns the best candidate name for a service: an exact match, else the
// unique candidate prefixed with "<service>-", else "".
func Match(service string, candidates []string) string {
	for _, c := range candidates {
		if c == service {
			return c
		}
	}
	var prefixed []string
	for _, c := range candidates {
		if strings.HasPrefix(c, service+"-") {
			prefixed = append(prefixed, c)
		}
	}
	if len(prefixed) == 1 {
		return prefixed[0]
	}
	return ""
}
