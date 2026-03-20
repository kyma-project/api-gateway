package externalgateway

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

const (
	externalRegionsConfigMapName = "external-gateway-regions"
	regionsYAMLKey               = "regions.yaml"
)

// RegionMetadata represents a single region entry from the ConfigMap
type RegionMetadata struct {
	Provider     string   `yaml:"Provider"`
	Region       string   `yaml:"Region"`
	CertSubjects []string `yaml:"CertSubjects"`
}

// RegionCertSubject represents parsed X509 certificate fields for a region
type RegionCertSubject struct {
	Region string   // e.g., "aws/eu-central-1"
	CN     string   // Common Name
	L      string   // Locality
	OU     []string // Organizational Units (multiple per region)
}

// extractField extracts a single X509 field value from a certificate subject string
// Example: extractField("CN=test, OU=org", "CN") returns "test"
func extractField(subject, field string) string {
	pattern := field + "=([^,]+)"
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(subject)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractAllFields extracts all values for a given X509 field from a certificate subject string
// Example: extractAllFields("OU=org1, OU=org2", "OU") returns ["org1", "org2"]
func extractAllFields(subject, field string) []string {
	pattern := field + "=([^,]+)"
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(subject, -1)

	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, strings.TrimSpace(match[1]))
		}
	}
	return results
}

// getRegionsYAMLFromConfigMap extracts regions YAML data from ConfigMap
// If ConfigMap has exactly one key, uses that key automatically
// If ConfigMap has multiple keys, looks for the expected regionsYAMLKey
func getRegionsYAMLFromConfigMap(configMap *corev1.ConfigMap, namespace, configMapName string) (string, error) {
	if len(configMap.Data) == 0 {
		return "", fmt.Errorf("ConfigMap %s/%s is empty", namespace, configMapName)
	}

	if len(configMap.Data) == 1 {
		for _, value := range configMap.Data {
			ctrl.Log.Info("Using the only available key from ConfigMap")
			return value, nil
		}
	}

	regionsYAML, exists := configMap.Data[regionsYAMLKey]
	if !exists {
		return "", fmt.Errorf("ConfigMap %s/%s does not contain '%s' key", namespace, configMapName, regionsYAMLKey)
	}
	return regionsYAML, nil
}

// ResolveRegionCertSubjects reads the ConfigMap specified in the ExternalGateway spec and extracts certificate subjects
// for the region specified in the ExternalGateway spec, parsing X509 fields from the certificate subject strings
func ResolveRegionCertSubjects(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway) ([]RegionCertSubject, error) {
	requestedRegion := external.Spec.Region
	configMapName := external.Spec.RegionsConfigMap

	ctrl.Log.Info("Resolving certificate subjects for region",
		"region", requestedRegion,
		"configMapName", configMapName,
		"namespace", external.Namespace)

	// Read ConfigMap from application namespace
	configMap := &corev1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{
		Name:      configMapName,
		Namespace: external.Namespace,
	}

	if err := k8sClient.Get(ctx, configMapNamespacedName, configMap); err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap %s/%s: %w", external.Namespace, configMapName, err)
	}

	// Get regions data from ConfigMap
	regionsYAML, err := getRegionsYAMLFromConfigMap(configMap, external.Namespace, configMapName)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var regions []RegionMetadata
	if err := yaml.Unmarshal([]byte(regionsYAML), &regions); err != nil {
		return nil, fmt.Errorf("failed to parse regions.yaml: %w", err)
	}

	// Build a map for quick lookup: "provider/region" -> []certSubjects
	regionMap := make(map[string][]string)
	for _, region := range regions {
		// Normalize to lowercase for case-insensitive matching
		key := fmt.Sprintf("%s/%s", strings.ToLower(region.Provider), strings.ToLower(region.Region))
		regionMap[key] = region.CertSubjects
	}

	// Get cert subjects for the requested region
	normalizedRegion := strings.ToLower(requestedRegion)
	subjects, exists := regionMap[normalizedRegion]
	if !exists {
		return nil, fmt.Errorf("region %s not found in ConfigMap %s/%s", requestedRegion, external.Namespace, configMapName)
	}

	// Parse each certificate subject string and extract X509 fields
	var certSubjects []RegionCertSubject
	for _, subject := range subjects {
		cn := extractField(subject, "CN")
		l := extractField(subject, "L")
		ou := extractAllFields(subject, "OU")

		certSubjects = append(certSubjects, RegionCertSubject{
			Region: normalizedRegion,
			CN:     cn,
			L:      l,
			OU:     ou,
		})
	}

	if len(certSubjects) == 0 {
		return nil, fmt.Errorf("no certificate subjects found for requested region: %v", requestedRegion)
	}

	ctrl.Log.Info("Resolved certificate subjects", "count", len(certSubjects), "region", requestedRegion)
	return certSubjects, nil
}
