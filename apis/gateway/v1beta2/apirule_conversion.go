package v1beta2

import (
	"encoding/json"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this ApiRule to the Hub version (v1beta1).
func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.APIRule)

	specData, err := json.Marshal(src.Spec)
	if err != nil {
		return err
	}

	err = json.Unmarshal(specData, &dst.Spec)
	if err != nil {
		return err
	}

	statusData, err := json.Marshal(src.Status)
	if err != nil {
		return err
	}

	err = json.Unmarshal(statusData, &dst.Status)
	if err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	annotations := dst.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["gateway.kyma-project.io/converted"] = "true"
	dst.SetAnnotations(annotations)

	// Only one host is supported in v1beta1, so we use the first one from the list
	hosts := src.Spec.Hosts
	dst.Spec.Host = hosts[0]

	for _, rule := range dst.Spec.Rules {
		srcRule := findBeta2Rule(src.Spec.Rules, &rule)
		// No Auth
		if srcRule.NoAuth != nil && *srcRule.NoAuth {
			rule.AccessStrategies = append(rule.AccessStrategies, &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name: "no_auth",
				},
			})
		}
	}

	return nil
}

// ConvertFrom converts this ApiRule from the Hub version (v1beta1).
func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.APIRule)
	specData, err := json.Marshal(src.Spec)
	if err != nil {
		return err
	}

	err = json.Unmarshal(specData, &dst.Spec)
	if err != nil {
		return err
	}

	statusData, err := json.Marshal(src.Status)
	if err != nil {
		return err
	}

	err = json.Unmarshal(statusData, &dst.Status)
	if err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta

	// Only one host is supported in v1beta1, so we use the first one from the list
	host := src.Spec.Host
	dst.Spec.Hosts = append(dst.Spec.Hosts, host)

	for _, dstRule := range dst.Spec.Rules {
		srcRule := findBeta1Rule(src.Spec.Rules, &dstRule)
		for _, accessStrategy := range srcRule.AccessStrategies {
			// No Auth
			if accessStrategy.Handler.Name == "no_auth" {
				dstRule.NoAuth = new(bool)
			}
		}
	}

	return nil
}

func findBeta1Rule(srcRules []v1beta1.Rule, dstRule *Rule) *v1beta1.Rule {
	for _, srcRule := range srcRules {
		if srcRule.Path == dstRule.Path && containsAllMethods(srcRule.Methods, dstRule.Methods) {
			return &srcRule
		}
	}

	return nil
}

func findBeta2Rule(srcRules []Rule, dstRule *v1beta1.Rule) *Rule {
	for _, srcRule := range srcRules {
		if srcRule.Path == dstRule.Path && containsAllMethods(dstRule.Methods, srcRule.Methods) {
			return &srcRule
		}
	}

	return nil
}

func containsAllMethods(srcMethods []v1beta1.HttpMethod, dstMethods []HttpMethod) bool {
	countMap1 := make(map[v1beta1.HttpMethod]int)
	countMap2 := make(map[HttpMethod]int)

	for _, str := range srcMethods {
		countMap1[str]++
	}
	for _, str := range dstMethods {
		countMap2[str]++
	}

	for method, count := range countMap1 {
		httpMethod := HttpMethod(method)
		if count != countMap2[httpMethod] {
			return false
		}
	}

	return len(srcMethods) == len(dstMethods)
}
