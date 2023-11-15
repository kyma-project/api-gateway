export function generateNamespaceName() {
    return generateRandomName("a-busola-test");
}

export function generateRandomName(name) {
    const random = Math.floor(Math.random() * 9999) + 1000;
    return `${name}-${random}`;
}