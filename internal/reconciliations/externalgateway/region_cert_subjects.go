package externalgateway

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

const (
	regionsYAMLKey = "regions.yaml"
)

// RegionMetadata represents a single region entry from the ConfigMap
type RegionMetadata struct {
	Name     string   `yaml:"name"`     // Region identifier, e.g., "eu10", "us10"
	IPs      []string `yaml:"ips"`      // List of IP addresses for the region
	Subjects []string `yaml:"subjects"` // Certificate subjects
}

// RegionsConfig is the root structure when parsing regions list
type RegionsConfig struct {
	Regions []RegionMetadata `yaml:"regions"`
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

// reverseSubjectParts reverses the order of certificate subject parts to match Envoy's format
// Input:  "C=US, O=Example Corp, OU=Clients, OU=UUID, L=lb, CN=aws/eu-central-1"
// Output: "CN=aws/eu-central-1, L=lb, OU=UUID, OU=Clients, O=Example Corp, C=US"
func reverseSubjectParts(input string) string {
	parts := strings.Split(input, ", ")
	slices.Reverse(parts)
	return strings.Join(parts, ",")
}

// ResolveRegionCertSubjects reads the ConfigMap specified in the ExternalGateway spec and extracts certificate subjects
// for the region specified in the ExternalGateway spec, returning reversed subject strings to match Envoy's format
func ResolveRegionCertSubjects(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway) ([]string, error) {
	requestedRegion := external.Spec.Region
	configMapName := external.Spec.RegionsConfigMap

	ctrl.Log.Info("Resolving certificate subjects for",
		"requestedRegion", requestedRegion,
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

	// Parse YAML with root "regions:" key
	var config RegionsConfig
	if err := yaml.Unmarshal([]byte(regionsYAML), &config); err != nil {
		return nil, fmt.Errorf("failed to parse regions.yaml: %w", err)
	}

	ctrl.Log.Info("Parsed regions from ConfigMap", "count", len(config.Regions))

	if len(config.Regions) == 0 {
		return nil, fmt.Errorf("no regions found in ConfigMap %s/%s", external.Namespace, configMapName)
	}

	// Build a map for quick lookup: "name" -> []subjects
	regionMap := make(map[string][]string)
	for _, region := range config.Regions {
		// Normalize region name to lowercase for case-insensitive matching
		key := strings.ToLower(region.Name)
		regionMap[key] = region.Subjects
	}

	// Get cert subjects for the requested region
	normalizedRegion := strings.ToLower(requestedRegion)
	subjects, exists := regionMap[normalizedRegion]
	if !exists {
		return nil, fmt.Errorf("requestedRegion %s not found in ConfigMap %s/%s", requestedRegion, external.Namespace, configMapName)
	}

	if len(subjects) == 0 {
		return nil, fmt.Errorf("no certificate subjects found for requested region: %v", requestedRegion)
	}

	// Reverse each certificate subject string to match Envoy's order
	var reversedSubjects []string
	for _, subject := range subjects {
		reversedSubjects = append(reversedSubjects, reverseSubjectParts(subject))
	}

	ctrl.Log.Info("Resolved certificate subjects for region", "count", len(reversedSubjects), "requestedRegion", requestedRegion)
	return reversedSubjects, nil
}
