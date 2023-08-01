package validation

import (
	"github.com/kyma-project/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("Timeout validation", func() {

	Context("for APIRule", func() {
		It("should return failure when timeout is greater than 65m", func() {
			// given
			timeout, err := time.ParseDuration("65m1s")
			Expect(err).ToNot(HaveOccurred())
			apiRule := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Timeout: &metav1.Duration{Duration: timeout},
				},
			}

			// when
			failures := timeoutValidator{}.validateApiRule(apiRule)

			// then
			Expect(failures).To(HaveLen(1))
			Expect(failures[0].AttributePath).To(Equal("spec.timeout"))
			Expect(failures[0].Message).To(Equal("Timeout must not exceed 65m"))
		})

		It("should return failure when timeout is 0", func() {
			// given
			timeout, err := time.ParseDuration("0")
			Expect(err).ToNot(HaveOccurred())
			apiRule := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Timeout: &metav1.Duration{Duration: timeout},
				},
			}

			// when
			failures := timeoutValidator{}.validateApiRule(apiRule)

			// then
			Expect(failures).To(HaveLen(1))
			Expect(failures[0].AttributePath).To(Equal("spec.timeout"))
			Expect(failures[0].Message).To(Equal("Timeout must not be 0 or lower"))
		})

		It("should return no failure when timeout is less than 65m", func() {
			// given
			timeout, err := time.ParseDuration("64m59s")
			Expect(err).ToNot(HaveOccurred())
			apiRule := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Timeout: &metav1.Duration{Duration: timeout},
				},
			}

			// when
			failures := timeoutValidator{}.validateApiRule(apiRule)

			// then
			Expect(failures).To(HaveLen(0))
		})

		It("should return no failure when timeout is nil", func() {
			// given
			apiRule := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{},
			}

			// when
			failures := timeoutValidator{}.validateApiRule(apiRule)

			// then
			Expect(failures).To(HaveLen(0))
		})

		It("should return no failure when timeout is 65m", func() {
			// given
			timeout, err := time.ParseDuration("65m")
			Expect(err).ToNot(HaveOccurred())
			apiRule := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Timeout: &metav1.Duration{Duration: timeout},
				},
			}

			// when
			failures := timeoutValidator{}.validateApiRule(apiRule)

			// then
			Expect(failures).To(HaveLen(0))
		})
	})

	Context("for Rule", func() {
		It("should return failure when timeout is greater than 65m", func() {
			// given
			timeout, err := time.ParseDuration("65m1s")
			Expect(err).ToNot(HaveOccurred())
			rule := v1beta1.Rule{
				Timeout: &metav1.Duration{Duration: timeout},
			}

			// when
			failures := timeoutValidator{}.validateRule(rule, "some.rule")

			// then
			Expect(failures).To(HaveLen(1))
			Expect(failures[0].AttributePath).To(Equal("some.rule.timeout"))
			Expect(failures[0].Message).To(Equal("Timeout must not exceed 65m"))
		})

		It("should return failure when timeout is 0", func() {
			// given
			timeout, err := time.ParseDuration("0")
			Expect(err).ToNot(HaveOccurred())
			rule := v1beta1.Rule{
				Timeout: &metav1.Duration{Duration: timeout},
			}

			// when
			failures := timeoutValidator{}.validateRule(rule, "some.rule")

			// then
			Expect(failures).To(HaveLen(1))
			Expect(failures[0].AttributePath).To(Equal("some.rule.timeout"))
			Expect(failures[0].Message).To(Equal("Timeout must not be 0 or lower"))
		})

		It("should return no failure when timeout is less than 65m", func() {
			// given
			timeout, err := time.ParseDuration("64m59s")
			Expect(err).ToNot(HaveOccurred())
			rule := v1beta1.Rule{
				Timeout: &metav1.Duration{Duration: timeout},
			}

			// when
			failures := timeoutValidator{}.validateRule(rule, "some.rule")

			// then
			Expect(failures).To(HaveLen(0))
		})

		It("should return no failure when timeout is nil", func() {
			// given
			rule := v1beta1.Rule{}

			// when
			failures := timeoutValidator{}.validateRule(rule, "some.rule")

			// then
			Expect(failures).To(HaveLen(0))
		})

		It("should return no failure when timeout is 65m", func() {
			// given
			timeout, err := time.ParseDuration("65m")
			Expect(err).ToNot(HaveOccurred())
			rule := v1beta1.Rule{
				Timeout: &metav1.Duration{Duration: timeout},
			}

			// when
			failures := timeoutValidator{}.validateRule(rule, "some.rule")

			// then
			Expect(failures).To(HaveLen(0))
		})

	})
})
