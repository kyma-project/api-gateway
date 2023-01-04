package v1alpha1

import (
	"testing"

	"github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test_ConvertTo(t *testing.T) {
	host := "some-host"
	alpha := APIRule{
		Spec: APIRuleSpec{
			Service: &Service{
				Host: &host,
			},
		},
	}
	beta := v1beta1.APIRule{}
	err := alpha.ConvertTo(&beta)
	It("Should convert with moving host to spec.host level", func() {
		Expect(err).To(BeNil())
		Expect(beta.Spec.Host).To(Equal(alpha.Spec.Service.Host))
	})
}

func Test_ConvertToNilAssert(t *testing.T) {
	alpha := APIRule{
		Spec: APIRuleSpec{
			Service: &Service{
				Host: nil,
			},
		},
	}
	beta := v1beta1.APIRule{}
	err := alpha.ConvertTo(&beta)
	It("Should fail if host is nil", func() {
		Expect(err).To(BeNil()) // Error will be nil as this is the prefered way for conversions to proceed
		Expect(beta.Spec.Host).To(BeNil())
	})

	alpha = APIRule{
		Spec: APIRuleSpec{
			Service: nil,
		},
	}
	err1 := alpha.ConvertTo(&beta)
	It("Should fail if service is nil", func() {
		Expect(err1).To(BeNil()) // Error will be nil as this is the prefered way for conversions to proceed
		Expect(beta.Spec).To(BeNil())
	})
}

func Test_ConvertFrom(t *testing.T) {
	host := "some-host"
	beta := v1beta1.APIRule{
		Spec: v1beta1.APIRuleSpec{
			Host: &host,
		},
	}

	alpha := APIRule{}
	err := alpha.ConvertFrom(&beta)
	It("Should convert with moving host to spec.service.host level", func() {
		Expect(err).To(BeNil())
		Expect(beta.Spec.Host).To(Equal(alpha.Spec.Service.Host))
	})
}

func Test_ConvertFromNilAssert(t *testing.T) {
	beta := v1beta1.APIRule{
		Spec: v1beta1.APIRuleSpec{
			Host: nil,
		},
	}
	alpha := APIRule{}
	err := alpha.ConvertTo(&beta)
	It("Should fail if host is nil", func() {
		Expect(err).To(BeNil()) // Error will be nil as this is the prefered way for conversions to proceed
		Expect(alpha.Spec.Service).To(BeNil())
	})
}
