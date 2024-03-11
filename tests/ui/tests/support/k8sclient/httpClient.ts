import config from "../../config";
import {getAuthHeaders} from "./auth";

export async function postApiEndpoint(url: string, data: Object): Promise<any> {
    return post(`${config.backendApiUrl}${url}`, data);
}

export async function postApisEndpoint(url: string, data: Object): Promise<any> {
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
    } catch (e) {
        cy.log(e);
        return false;
    }
}

export async function deleteResourceApiEndpoint(resourceUrl: string): Promise<boolean> {
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
    } catch (e) {
        cy.log(e);
        return false;
    }
}

