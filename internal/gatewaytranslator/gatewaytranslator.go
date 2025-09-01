package gatewaytranslator

import (
	"fmt"
	"slices"
	"strings"
)

func TranslateGatewayNameToNewFormat(gatewayName string, namespace string) (string, error) {
	splitGatewayName := strings.Split(gatewayName, ".")
	switch len(splitGatewayName) {
	case 5: // old format with .svc.cluster.local suffix
		gatewayName = strings.TrimSuffix(gatewayName, ".svc.cluster.local")
	case 4: // old format with .svc.cluster suffix
		gatewayName = strings.TrimSuffix(gatewayName, ".svc.cluster")
	case 3: // old format with .svc suffix
		gatewayName = strings.TrimSuffix(gatewayName, ".svc")
	case 2: // old format with no suffix
		return fmt.Sprintf("%s/%s", splitGatewayName[1], splitGatewayName[0]), nil
	case 1: // old format without namespace
		return fmt.Sprintf("%s/%s", namespace, gatewayName), nil
	}
	parts := strings.Split(gatewayName, ".")
	if len(parts) == 2 {
		return fmt.Sprintf("%s/%s", parts[1], parts[0]), nil
	}
	if len(parts) == 1 {
		return fmt.Sprintf("%s/%s", namespace, parts[0]), nil
	}
	return "", fmt.Errorf("gateway name (%s) is not in old gateway format", gatewayName)
}

func IsOldGatewayNameFormat(gatewayName string) bool {
	parts := strings.Split(gatewayName, ".")
	if len(parts) > 2 {
		suffix := strings.Join(parts[2:], ".")
		oldGatewayNameSuffixes := []string{"svc.cluster.local", "svc.cluster", "svc"}
		return slices.Contains(oldGatewayNameSuffixes, suffix)
	}
	if len(parts) <= 2 && !strings.Contains(gatewayName, "/") {
		return true
	}

	return false
}
