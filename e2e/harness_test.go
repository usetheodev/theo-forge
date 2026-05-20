//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/usetheodev/theo-forge/client"
	"github.com/usetheodev/theo-forge/model"
)

// --- Environment ---

const (
	defaultNamespace   = "argo"
	defaultServerHost  = "http://localhost:2746"
	defaultWaitTimeout = 180 * time.Second
)

// e2eEnv captures everything a test needs to talk to the cluster.
// Built once per process by getEnv; reused across tests.
type e2eEnv struct {
	kubeconfig string
	tokenFile  string
	token      string
	namespace  string
	host       string
}

// getEnv loads the E2E configuration from environment variables.
// Tests call requireEnv(t) instead — getEnv is exported only for
// other helpers in this package.
func getEnv() (*e2eEnv, error) {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		return nil, errors.New("KUBECONFIG not set — run `make e2e-up` first")
	}
	if _, err := os.Stat(kc); err != nil {
		return nil, fmt.Errorf("kubeconfig %s: %w", kc, err)
	}

	tokenFile := os.Getenv("ARGO_TOKEN_FILE")
	if tokenFile == "" {
		return nil, errors.New("ARGO_TOKEN_FILE not set — run `make e2e-up` first")
	}
	tokBytes, err := os.ReadFile(tokenFile) //nolint:gosec // path comes from test env, not user input
	if err != nil {
		return nil, fmt.Errorf("read token file %s: %w", tokenFile, err)
	}
	token := strings.TrimSpace(string(tokBytes))
	if token == "" {
		return nil, fmt.Errorf("token file %s is empty", tokenFile)
	}

	ns := os.Getenv("ARGO_NAMESPACE")
	if ns == "" {
		ns = defaultNamespace
	}
	host := os.Getenv("ARGO_HOST")
	if host == "" {
		host = defaultServerHost
	}
	return &e2eEnv{
		kubeconfig: kc,
		tokenFile:  tokenFile,
		token:      token,
		namespace:  ns,
		host:       host,
	}, nil
}

// requireEnv returns the e2e environment or fails the test loudly.
func requireEnv(t *testing.T) *e2eEnv {
	t.Helper()
	env, err := getEnv()
	if err != nil {
		t.Fatalf("E2E environment not ready: %v", err)
	}
	return env
}

// --- Clients ---

// argoClient returns a theo-forge REST client bound to the live argo-server.
// VerifySSL=false because port-forward terminates TLS at localhost without
// a valid cert in the test env.
func argoClient(t *testing.T) *client.WorkflowsService {
	t.Helper()
	env := requireEnv(t)
	svc := client.NewWorkflowsService(env.host, env.token, env.namespace)
	svc.VerifySSL = false
	return svc
}

// --- kubectl shim ---

// kubectl runs `kubectl ARGS...` against the e2e kubeconfig and returns
// stdout. Fails the test on non-zero exit code, surfacing both stdout
// and stderr in the message.
func kubectl(t *testing.T, args ...string) string {
	t.Helper()
	env := requireEnv(t)
	full := append([]string{"--kubeconfig", env.kubeconfig}, args...)
	cmd := exec.Command("kubectl", full...) //nolint:gosec // args come from the test, not user input
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("kubectl %s failed: %v\nstdout:\n%s\nstderr:\n%s",
			strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

// --- Workflow lifecycle helpers ---

// uniqueName produces a workflow name unique to this test run, so
// reruns and parallel tests do not collide on the cluster.
func uniqueName(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

// waitWorkflowSucceeded blocks until the workflow reaches a terminal
// phase or the timeout fires. Returns the final WorkflowModel for asserts.
func waitWorkflowSucceeded(t *testing.T, name string, timeout time.Duration) model.WorkflowModel {
	t.Helper()
	svc := argoClient(t)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		wf, err := svc.GetWorkflow(ctx, name, "")
		cancel()
		if err == nil && wf.Status != nil {
			switch wf.Status.Phase {
			case model.WorkflowSucceeded:
				return wf
			case model.WorkflowFailed, model.WorkflowError, model.WorkflowTerminated:
				yaml := kubectl(t, "-n", "argo", "get", "wf", name, "-o", "yaml")
				t.Fatalf("workflow %s ended in %s — full manifest:\n%s",
					name, wf.Status.Phase, yaml)
			}
		}
		time.Sleep(2 * time.Second)
	}
	yaml := kubectl(t, "-n", "argo", "get", "wf", name, "-o", "yaml")
	t.Fatalf("workflow %s did not reach Succeeded in %s — manifest:\n%s",
		name, timeout, yaml)
	return model.WorkflowModel{}
}

// cleanupWorkflow deletes the workflow on test exit. Best-effort —
// errors during teardown do not fail the test.
func cleanupWorkflow(t *testing.T, name string) {
	t.Helper()
	t.Cleanup(func() {
		env, err := getEnv()
		if err != nil {
			return
		}
		cmd := exec.Command("kubectl", "--kubeconfig", env.kubeconfig, //nolint:gosec
			"-n", env.namespace, "delete", "wf", name, "--ignore-not-found")
		_ = cmd.Run()
	})
}

// dumpWorkflow returns the cluster's view of the workflow as YAML.
// Use in assertion failure messages to surface what Argo actually sees.
func dumpWorkflow(t *testing.T, name string) string {
	t.Helper()
	env := requireEnv(t)
	out := kubectl(t, "-n", env.namespace, "get", "wf", name, "-o", "yaml")
	return out
}

// writeWorkflowYAML writes a workflow YAML into the cluster via
// `kubectl apply`. Returns the workflow's effective name (Name or
// the generated name from .metadata.generateName).
func writeWorkflowYAML(t *testing.T, name, yamlStr string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name+".yaml")
	if err := os.WriteFile(path, []byte(yamlStr), 0o600); err != nil {
		t.Fatalf("write temp manifest: %v", err)
	}
	kubectl(t, "-n", "argo", "apply", "-f", path)
}
