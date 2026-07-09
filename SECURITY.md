# Security Policy

## Supported versions

`pier` is distributed as a single binary with no long-term support branches.
Only the latest released version receives security fixes.

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Report privately through GitHub's [private vulnerability
reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
(the **Security → Report a vulnerability** tab of the repository). If that is
unavailable, email the maintainer at <secontreras@comparaonline.com>.

Please include reproduction steps and the affected version (`pier --version`).
You can expect an initial acknowledgement within a few business days.

## Handling of secrets

`pier` is a `kubectl` wrapper and never stores credentials of its own — it
inherits your existing kubeconfig authentication.

One behavior is worth calling out explicitly: **`pier secret get --decode`
prints Secret values in cleartext to your terminal**, exactly as `kubectl get
secret -o jsonpath` would. This is intended, but be mindful of shell history,
terminal scrollback, and screen sharing when using it.
