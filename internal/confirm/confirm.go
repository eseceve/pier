// Package confirm implements pier's environment-forward guardrail prompt for
// mutating operations against protected environments.
package confirm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/comparaonline/pier/internal/config"
)

// Prompt shows the environment, resolved coordinates and the exact action, then
// requires the user to type the service name. It returns an error if the typed
// value does not match.
func Prompt(in io.Reader, out io.Writer, t config.Target, action string) error {
	if _, err := fmt.Fprintf(out, "⚠  %s — context: %s, namespace: %s\n", strings.ToUpper(t.Env), t.Context, t.Namespace); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "   This will: %s\n", action); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "   Type the service name to confirm (%s): ", t.Service); err != nil {
		return err
	}

	line, _ := bufio.NewReader(in).ReadString('\n')
	if strings.TrimSpace(line) != t.Service {
		return errors.New("aborted: confirmation did not match the service name")
	}
	return nil
}
