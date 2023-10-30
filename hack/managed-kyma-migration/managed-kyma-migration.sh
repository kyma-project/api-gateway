#!/usr/bin/env bash

set -eo pipefail

function check_apigateway_status () {
	local number=1
	while [[ $number -le 100 ]] ; do
		echo ">--> checking kyma status #$number"
		local STATUS=$(kubectl get apigateway default -o jsonpath='{.status.state}')
		echo "apigateway status: ${STATUS:='UNKNOWN'}"
		[[ "$STATUS" == "Ready" ]] && return 0
		sleep 5
        	((number = number + 1))
	done

	kubectl get all --all-namespaces
	exit 1
}

# Verify API Gateway module template is available on cluster
api_gateway_module_template_count=$(kubectl get moduletemplates.operator.kyma-project.io -A --output json | jq '.items | map(. | select(.spec.data.kind=="APIGateway")) | length')

if [ "$api_gateway_module_template_count" -eq 0 ]; then
  echo "API Gateway module template is not available on cluster"
  exit 1
fi

# Fetch Kyma CR name managed by lifecycle-manager
kyma_cr_name=$(kubectl get kyma -n kyma-system --no-headers -o custom-columns=":metadata.name")

# Fetch all modules, if modules is not defined, fallback to an empty array and count the number modules that have the name "api-gateway"
api_gateway_module_count=$(kubectl get kyma "$kyma_cr_name" -n kyma-system -o json | jq '.spec.modules | if . == null then [] else . end | map(. | select(.name=="api-gateway")) | length')

if [  "$api_gateway_module_count" -gt 0 ]; then
  echo "API Gateway module already present on Kyma CR, skipping migration"
  kubectl delete deployment -n kyma-system api-gateway
  exit 0
fi

kyma_cr_modules=$(kubectl get kyma "$kyma_cr_name" -n kyma-system -o json | jq '.spec.modules')
if [ "$kyma_cr_modules" == "null" ]; then
  echo "No modules defined on Kyma CR yet, initializing modules array and adding api-gateway module"
  kubectl patch kyma "$kyma_cr_name" -n kyma-system --type='json' -p='[{"op": "add", "path": "/spec/modules", "value": [{"name": "api-gateway"}] }]'
else
  echo "Proceeding with migration by adding API Gateway module to Kyma CR $kyma_cr_name"
  kubectl patch kyma "$kyma_cr_name" -n kyma-system --type='json' -p='[{"op": "add", "path": "/spec/modules/-", "value": {"name": "api-gateway"} }]'
fi

check_apigateway_status

#Removing api-gateway deployment so it won't interfere with APIRules reconciliation
kubectl delete deployment -n kyma-system api-gateway

echo "API Gateway CR migration completed successfully"
exit 0
