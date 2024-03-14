import config from "../dashboard";
import {getAuthHeaders} from "./kubeconfig";

export async function postApi(url: string, data: Object): Promise<boolean> {
    return post(`${config.backendApiUrl}${url}`, data);
}

export async function postApis(url: string, data: Object): Promise<boolean> {
    return post(`${config.backendApisUrl}${url}`, data);
}

export async function post(url: string, data: Object): Promise<boolean> {
    try {
        const authHeaders = await getAuthHeaders();
        const response = await fetch(url,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Accept: '*/*',
                    ...authHeaders,
                },
                credentials: "same-origin",
                body: JSON.stringify(data),

            });
        return response.ok;
    } catch (e: any) {
        cy.log(e);
        return false;
    }
}

export async function deleteResource(resourceUrl: string): Promise<boolean> {
    try {
        const authHeaders = await getAuthHeaders();
        const response = await fetch(`${config.backendApiUrl}${resourceUrl}`,
            {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    Accept: '*/*',
                    ...authHeaders,
                },
                credentials: "same-origin",
            });
        return response.ok;
    } catch (e: any) {
        cy.log(e);
        return false;
    }
}

