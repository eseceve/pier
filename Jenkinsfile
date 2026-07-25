// pier CI/CD — Node-free Go pipeline (svu + GoReleaser).
//
// DRAFT: this pipeline has NOT been run against real Jenkins yet. Verify before
// relying on it:
//   * the agent can run Docker and pull the `golang:1.26` image (which bundles git);
//   * the `GitHubJenkinsAccessToken` credential exists and can push tags and create
//     releases on eseceve/pier;
//   * git operations inside the container may need `safe.directory` on the mounted
//     workspace depending on uid mapping;
//   * `jenkinsNotification()` and automatic SCM checkout rely on the org's global
//     Jenkins shared library / job config (same as ai-funnel-webapp);
//   * `git push` uses the agent's git credentials for the remote (not GITHUB_TOKEN);
//   * branch model: `develop` => `beta` prerelease, `main` => stable release;
//   * svu v3's `next`/`current` output format (one version per line, nothing on
//     stdout from `go install`) — release() parses stdout line-by-line and decides
//     "release warranted" from the stable next vs current, so confirm the exact
//     pre-release progression on develop against a real svu run.
pipeline {
  agent any

  options {
    timeout(time: 20, unit: 'MINUTES')
  }

  environment {
    APP_NAME = "pier"
    GO_IMAGE = "golang:1.26"
    GITHUB_TOKEN = credentials("GitHubJenkinsAccessToken")
  }

  stages {
    // Runs on every branch, including main/develop, so a release is never cut
    // without lint + tests + build passing first (stages run in order).
    stage("Test") {
      steps {
        script {
          go_run("go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && golangci-lint run")
          go_run("go test ./... -race")
          go_run("go build ./...")
        }
      }
    }

    stage("Release") {
      when { anyOf { branch "main"; branch "develop" } }
      steps {
        script { release() }
      }
    }
  }

  post {
    always {
      script {
        jenkinsNotification()
      }
    }
  }
}

// go_run executes a command inside the Go toolchain container against the workspace.
def go_run(cmd) {
  sh "docker run --rm -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c '${cmd}'"
}

// go_out is like go_run but returns the command's trimmed stdout.
def go_out(cmd) {
  return sh(
    script: "docker run --rm -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c '${cmd}'",
    returnStdout: true
  ).trim()
}

def release() {
  sh 'git config --global user.email "jenkins@comparaonline.com"'
  sh 'git config --global user.name "Jenkins"'

  def svuInstall = "go install github.com/caarlos0/svu/v3@latest"

  // Whether a release is warranted is decided like-for-like: the *stable* next
  // version vs the current tag. Installing svu once and printing both keeps it to
  // a single ephemeral container. The beta suffix is orthogonal formatting and
  // must not enter this comparison, or develop would always look "ahead".
  def probe = go_out("${svuInstall} && svu current && svu next")
  def lines = probe.split("\n")
  def current = lines[0].trim()
  def nextStable = lines.length > 1 ? lines[1].trim() : current

  if (nextStable == current) {
    echo "No release warranted (svu next == current: ${current})"
    return
  }

  def next = env.BRANCH_NAME == 'develop'
    ? go_out("${svuInstall} && svu next --pre-release beta")
    : nextStable

  sh "git tag ${next}"
  sh "git push origin ${next}"

  // GoReleaser builds the cross-platform binaries and publishes the GitHub Release.
  // GITHUB_TOKEN is forwarded as an env var (never interpolated into the command).
  // If it fails after the tag is pushed, delete the orphaned tag so the next build
  // re-computes a release instead of seeing next == current and silently skipping.
  try {
    sh "docker run --rm -e GITHUB_TOKEN -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c 'go install github.com/goreleaser/goreleaser/v2@latest && goreleaser release --clean'"
  } catch (err) {
    sh "git push origin :refs/tags/${next} || true"
    sh "git tag -d ${next} || true"
    throw err
  }

  if (env.BRANCH_NAME == 'main') {
    sh "git checkout develop || git checkout -b develop origin/develop"
    // Plain merge: a genuine conflict must fail the build for a human to resolve.
    // -X ours would auto-resolve by discarding main-only changes (e.g. a hotfix),
    // silently regressing them on the next develop -> main promotion.
    sh "git merge --no-edit origin/main"
    sh "git push origin develop"
  }
}
