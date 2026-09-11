#!/usr/bin/env bash

# Description: This script downloads k3d CLI and provisions a k3d cluster
# Environment variables:
# - K3D_CONFIGURATION - configuration preset, used to load configurations/${K3D_CONFIGURATION}/vars.sh

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

echo "Started provision script"

require_vars K3D_CONFIGURATION
load_configuration "${K3D_CONFIGURATION}"

K3S_IMAGE="rancher/k3s:v${KUBERNETES_VERSION}-k3s1"

echo "Configuration:"
echo "  Kubernetes version: ${KUBERNETES_VERSION}"
echo "  k3d version: ${K3D_VERSION}"
echo "  k3s image: ${K3S_IMAGE}"
echo "  Agents: ${AGENTS}"
echo "  Servers: ${SERVERS}"
echo "  Servers memory: ${SERVERS_MEMORY}g"
echo "  Use Calico: ${USE_CALICO}"
echo "  Calico version: ${CALICO_VERSION}"
echo "  Use KWOK: ${USE_KWOK}"
echo "  KWOK version: ${KWOK_VERSION}"
echo "  Kwok nodes: ${KWOK_NODES}"

# Function to install k3d
install_k3d() {
    if command -v k3d &> /dev/null; then
        echo "k3d is already installed: $(k3d version | head -1)"
        return
    fi

    curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

    echo "k3d installed successfully: $(k3d version | head -1)"
}

# Function to provision cluster with Calico
provision_calico_cluster() {
    if [ "${AGENTS}" -gt 0 ]; then
        echo "Error: script does not support calico setup with AGENTS > 0. Please use AGENTS=0."
        exit 1
    fi

    echo "Provisioning k3d cluster with Calico CNI..."

    k3d cluster create \
        --agents "${AGENTS}" \
        --servers "${SERVERS}" \
        --servers-memory "${SERVERS_MEMORY}g" \
        --port 80:80@loadbalancer \
        --port 443:443@loadbalancer \
        --k3s-arg "--flannel-backend=none@all" \
        --k3s-arg "--disable=traefik@server:*" \
        --k3s-arg '--tls-san=host.docker.internal@server:*' \
        --image "${K3S_IMAGE}"

    echo "Installing Calico ${CALICO_VERSION}..."
    kubectl create -f "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VERSION}/manifests/operator-crds.yaml"
    kubectl create -f "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VERSION}/manifests/tigera-operator.yaml"
    kubectl create -f "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VERSION}/manifests/custom-resources.yaml"

    kubectl rollout status -n kube-system deployment coredns
    kubectl patch installation default --type=merge -p '{"spec":{"cni":{"binDir":"/var/lib/rancher/k3s/data/cni", "confDir":"/var/lib/rancher/k3s/agent/etc/cni/net.d"}}}'

}

# Function to provision regular cluster (without traefik)
provision_regular_cluster() {
    echo "Provisioning k3d cluster (regular, without traefik)..."

    k3d cluster create \
        --agents "${AGENTS}" \
        --servers-memory "${SERVERS_MEMORY}g" \
        --port 80:80@loadbalancer \
        --port 443:443@loadbalancer \
        --k3s-arg '--disable=traefik@server:*' \
        --image "${K3S_IMAGE}"
}

setup_kwok() {
    echo "Installing KWOK"
    # KWOK repository
    KWOK_REPO=kubernetes-sigs/kwok
    kubectl apply -f "https://github.com/${KWOK_REPO}/releases/download/${KWOK_VERSION}/kwok.yaml"
    kubectl apply -f "https://github.com/${KWOK_REPO}/releases/download/${KWOK_VERSION}/stage-fast.yaml"
    kubectl apply -f "hack/manifests/chaos/job-pod-running.yaml"

    if [[ "${KWOK_NODES}" -gt 0 ]]; then
        check_envsubst_vars "${script_dir}/kwok-node-template.yaml"
        echo "Creating ${KWOK_NODES} fake Nodes..."
        for i in $(seq 1 "${KWOK_NODES}"); do
            KWOK_NODE_NAME="kwok-node-${i}" envsubst < "${script_dir}/kwok-node-template.yaml" | kubectl apply -f -
        done
    fi
}

echo "Install k3d"
install_k3d

echo "Provision cluster"
if [ "${USE_CALICO}" = true ]; then
    provision_calico_cluster
else
    provision_regular_cluster
fi

if [ "${USE_KWOK}" = true ]; then
    echo "Setup KWOK"
    setup_kwok
fi

echo "Provision script finished"
