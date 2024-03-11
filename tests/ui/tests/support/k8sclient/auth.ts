import {loadFixture} from "./loadFile";

export type KubernetesConfig = {
    clusters: {
        cluster: {
            server: string;
            "certificate-authority-data": string;
        };
    }[];
    "current-context": string;
    users: {
        user: {
            exec: {
                args: string[];
            };
        };
    }[];
};

export async function getAuthHeaders() {
    const cfg = await loadKubeconfig();
    return {
        'X-Cluster-Url': cfg.clusters[0].cluster.server,
        'X-Cluster-Certificate-Authority-Data': cfg.clusters[0].cluster["certificate-authority-data"],
        'X-Client-Certificate-Data': cfg.users[0].user["client-certificate-data"],
        'X-Client-Key-Data': cfg.users[0].user["client-key-data"],
    };
}

export async function loadKubeconfig() {
    return  await loadFixture('kubeconfig.yaml') as KubernetesConfig;
}

export async function getK8sCurrentContext() {
    const cfg = await loadKubeconfig();
    return cfg["current-context"];
}