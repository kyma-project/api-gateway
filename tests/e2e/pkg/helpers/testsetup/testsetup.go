package testsetup

import (
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2mock"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"testing"
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

	t.Logf("Creating namespace %s", ns)
	return testId, ns, infrastructure.CreateNamespace(t, ns, opts.NamespaceCreationOptions...)
}

type TestBackground struct {
	TestName          string
	Namespace         string
	TargetServiceName string
	TargetServicePort int
	Mock              *oauth2mock.Mock
}

// SetupRandomNamespaceWithOauth2MockAndHttpbin creates a namespace with a random ID,
// deploys the oauth2mock and httpbin services, and returns a TestBackground struct.
// It also sets up the namespace with sidecar injection enabled.
func SetupRandomNamespaceWithOauth2MockAndHttpbin(t *testing.T, options ...Option) (bckg TestBackground, err error) {
	t.Helper()
	options = append(options, WithSidecarInjectionEnabled())

	testId, ns, err := CreateNamespaceWithRandomID(t, options...)
	require.NoError(t, err, "Failed to create a test namespace")

	svcName, svcPort, err := httpbin.DeployHttpbin(t, httpbin.WithNamespace(ns))
	require.NoError(t, err, "Failed to deploy httpbin service")

	mock, err := oauth2mock.DeployMock(t, oauth2mock.WithNamespace(ns))
	require.NoError(t, err, "Failed to deploy oauth2mock")

	return TestBackground{
		TestName:          testId,
		Namespace:         ns,
		TargetServiceName: svcName,
		TargetServicePort: svcPort,
		Mock:              mock,
	}, nil
}
