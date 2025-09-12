package testsetup

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	oauth2mock "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2/mock"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
)

type Options struct {
	Prefix                   string
	NamespaceCreationOptions []infrastructure.NamespaceOption
}

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

func WithSidecarInjectionEnabled() Option {
	return func(o *Options) {
		o.NamespaceCreationOptions = append(o.NamespaceCreationOptions, infrastructure.WithSidecarInjectionEnabled())
	}
}

type Option func(*Options)

func CreateNamespaceWithRandomID(t *testing.T, options ...Option) (testId string, namespaceName string, err error) {
	t.Helper()
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	testId = envconf.RandomName("test", 16)
	ns := testId
	if opts.Prefix != "" {
		ns = opts.Prefix + "-" + testId
	}
	setup.DeclareCleanup(t,
		func() {
			if err := infrastructure.DeleteNamespace(t, ns); err != nil {
				t.Logf("Failed to delete namespace %s: %v", namespaceName, err)
			} else {
				t.Logf("Namespace %s deleted successfully", namespaceName)
			}
		},
	)
	t.Logf("Creating namespace %s", ns)
	return testId, ns, infrastructure.CreateNamespace(t, ns, opts.NamespaceCreationOptions...)
}

type TestBackground struct {
	TestName          string
	Namespace         string
	TargetServiceName string
	TargetServicePort int
	Provider          oauth2.Provider
}

// SetupRandomNamespaceWithOauth2MockAndHttpbin creates a namespace with a random ID,
// deploys the oauth2mock and httpbin services, and returns a TestBackground struct.
// It also sets up the namespace with sidecar injection enabled.
func SetupRandomNamespaceWithOauth2MockAndHttpbin(t *testing.T, options ...Option) (bckg TestBackground, err error) {
	t.Helper()
	options = append(options, WithSidecarInjectionEnabled())

	testId, ns, err := CreateNamespaceWithRandomID(t, options...)
	require.NoError(t, err, "Failed to create a test namespace")

	svcName, svcPort, err := httpbin.DeployHttpbin(t, ns)
	require.NoError(t, err, "Failed to deploy httpbin service")

	mock, err := oauth2mock.DeployMock(t, ns)
	require.NoError(t, err, "Failed to deploy oauth2mock")

	return TestBackground{
		TestName:          testId,
		Namespace:         ns,
		TargetServiceName: svcName,
		TargetServicePort: svcPort,
		Provider:          mock,
	}, nil
}
