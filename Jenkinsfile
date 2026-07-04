// pier CI/CD — Node-free Go pipeline (svu + GoReleaser).
//
// DRAFT: this pipeline has NOT been run against real Jenkins yet. Verify before
// relying on it:
//   * the agent can run Docker and pull the `golang:1.26` image (which bundles git);
//   * the `GitHubJenkinsAccessToken` credential exists and can push tags and create
//     releases on comparaonline/pier;
//   * git operations inside the container may need `safe.directory` on the mounted
//     workspace depending on uid mapping;
//   * `jenkinsNotification()` and automatic SCM checkout rely on the org's global
//     Jenkins shared library / job config (same as ai-funnel-webapp);
//   * `git push` uses the agent's git credentials for the remote (not GITHUB_TOKEN);
//   * branch model: `develop` => `beta` prerelease, `main` => stable release.
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
    stage("Test") {
      when { not { anyOf { branch "main"; branch "develop" } } }
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

  // svu is installed in each container invocation (containers are ephemeral, so a
  // tool installed in one go_out call is not present in the next).
  def prerelease = env.BRANCH_NAME == 'develop' ? '--pre-release beta' : ''
  def next = go_out("go install github.com/caarlos0/svu/v3@latest && svu next ${prerelease}")
  def current = go_out("go install github.com/caarlos0/svu/v3@latest && svu current")

  if (next == current) {
    echo "No release warranted (svu next == current: ${current})"
    return
  }

  sh "git tag ${next}"
  sh "git push origin ${next}"

  // GoReleaser builds the cross-platform binaries and publishes the GitHub Release.
  // GITHUB_TOKEN is forwarded as an env var (never interpolated into the command).
  sh "docker run --rm -e GITHUB_TOKEN -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c 'go install github.com/goreleaser/goreleaser/v2@latest && goreleaser release --clean'"

  if (env.BRANCH_NAME == 'main') {
    sh "git checkout develop || git checkout -b develop origin/develop"
    sh "git merge --strategy-option=ours origin/main"
    sh "git push origin develop"
  }
}
