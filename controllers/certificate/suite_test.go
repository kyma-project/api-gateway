/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package certificate_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	gikgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

const (
	testNamespace     = "kyma-system"
	eventuallyTimeout = time.Second * 20
)

var (
	ctx context.Context
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "API Gateway Certificate Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report gikgotypes.Report) {
	tests.GenerateGinkgoJunitReport("api-gateway-certificate-suite", report)
})

func createFakeClient(objects ...client.Object) client.Client {
	fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).WithObjects(objects...).WithStatusSubresource(objects...).Build()

	return fakeClient
}

func getTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(appsv1.AddToScheme(s))
	utilruntime.Must(apiextensionsv1.AddToScheme(s))
	utilruntime.Must(admissionregistrationv1.AddToScheme(s))
	return s
}
