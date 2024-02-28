Cypress.on('uncaught:exception', err => {
    Cypress.log(err);
    if (
        err.message.includes('Unexpected usage') ||
        err.message.includes(
            "Cannot read properties of undefined (reading 'uri')",
        ) ||
        err.message.includes('ResizeObserver loop limit exceeded') ||
        err.message.includes('502 Bad Gateway') ||
        err.message.includes(
            'ResizeObserver loop completed with undelivered notifications',
        ) ||
        err.message.includes(
            "Cannot read properties of null (reading 'sendError')",
        ) ||
        err.message.includes(
            "Uncaught NetworkError: Failed to execute 'importScripts' on 'WorkerGlobalScope': The script at",
        ) ||
        err.message.includes(
            "Cannot read properties of undefined (reading 'category')",
        ) ||
        err.message.includes(
            "Cannot read properties of null (reading 'querySelector')",
        ) ||
        err.message.includes(
            "Cannot read properties of undefined (reading 'hasAttribute')",
        )
    ) {
        return false;
    }

});
