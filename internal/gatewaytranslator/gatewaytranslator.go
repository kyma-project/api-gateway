package gatewaytranslator

import (
	"fmt"
	"strings"
)

const oldGatewayNameSuffix = ".svc.cluster.local"

func TranslateGatewayNameToNewFormat(gatewayName string) (string, error) {
	s := strings.TrimSuffix(gatewayName, oldGatewayNameSuffix)
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("gateway name (%s) is not in old gateway format", gatewayName)
	}
	return fmt.Sprintf("%s/%s", parts[1], parts[0]), nil
}

func IsOldGatewayNameFormat(gatewayName string) bool {

	if !strings.HasSuffix(gatewayName, oldGatewayNameSuffix) {
		return false
	}
	s := strings.TrimSuffix(gatewayName, oldGatewayNameSuffix)
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return false
	}
	return true
}
