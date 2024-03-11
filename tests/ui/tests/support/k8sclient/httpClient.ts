import config from "../../config";
import {getAuthHeaders} from "./auth";


export async function post(apiUrl: string, data: Object): Promise<boolean> {
    try {
        const authHeaders = await getAuthHeaders();
        const response = await fetch(`${config.backendApiUrl}${apiUrl}`,
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
    } catch (e) {
        cy.log(e);
        return false;
    }
}

