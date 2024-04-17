export function generateNamespaceName() : string {
    return generateRandomName("a-busola-test");
}

export function generateRandomName(name: string) : string {
    const random = Math.floor(Math.random() * 9999) + 1000;
    return `${name}-${random}`;
}