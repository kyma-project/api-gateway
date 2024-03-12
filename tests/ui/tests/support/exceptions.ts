const ignorableErrors = [
    "Unexpected usage",
    "Cannot read properties of undefined (reading 'uri')",
    "ResizeObserver loop limit exceeded",
    "502 Bad Gateway",
    "ResizeObserver loop completed with undelivered notifications",
    "Cannot read properties of null (reading 'sendError')",
    "Uncaught NetworkError: Failed to execute 'importScripts' on 'WorkerGlobalScope': The script at",
    "Cannot read properties of undefined (reading 'category')",
    "Cannot read properties of null (reading 'querySelector')",
    "Cannot read properties of undefined (reading 'hasAttribute')",
];

function isIgnorableError(err: Error): boolean {
    return ignorableErrors.some(errorMessage => err.message.includes(errorMessage));
}
Cypress.on('uncaught:exception', (err : Error): boolean => {
    Cypress.log(err);

    return !isIgnorableError(err);
});

