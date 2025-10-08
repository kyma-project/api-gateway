# Mutual TLS Authentication
Learn what mutual TLS (mTLS) is, how it works, and how to implement it in SAP BTP, Kyma runtime.

## What Is Mutual TLS?
mTLS is a security protocol that ensures that both the client and the server authenticate each other. This two-way authentication ensures a higher level of security compared to simple TLS, where only the client checks the server's identity.

TLS brings the following benefits:

- It reduces the risk of a man-in-the-middle attacks, which occur when an attacker intercepts communication between two parties without their knowledge. The encryption ensures that even if someone retrieves the data, they cannot decrypt or modify it.

- The client is able to verify that the server is actually who they claim to be. For example, when opening SAP Help Portal (the server) your browser (the client) ensures that the certificate that the website presents is valid, confirming you are accessing the official website managed by SAP.

Additionally, with mTLS, the server is able to verify that the client is actually who they claim to be. For example, if you open an internal company website (the client), the application (the server) might request your certificate and only continue the session after it confirms your identity.

## Understanding How mTLS Works
Both TLS and mTLS protocoles are based on the concept of asymmetric encryption. Asymmetric encryption involves the use of two keys:
- The private key, which is kept secret
- The public key, which is shared publicly available

For example, imagine a server that has both a private and a public key. A client encrypts a message using the server's public key and then sends it to the server. Only the server, with its corresponding private key, can decrypt this message. This demonstrates the server's identity because only the genuine owner of the private key could perform the decryption.

In practice, however, it's not possible to collect, manage, and verify the authenticity of many public keys presented by each application you interact with. To resolve the problem, communication protocols use the public-key infrastructure, in which one or more third-parties, known as certificate authorities (CAs), certify ownership of key pairs by issuing digital certificates. A digital certificate includes such information as:
- The public key
- The details on the owner of the certificate
- A signature of the authority that confirms that the certificate's contents are valid
- The certificate's validity period

Certificates are typically signed by other certificates, creating trust chains where each certificate is signed by the one above it. If you trust the top-most (root) certificate in a chain, you can trust all the certificates below it.

## Using mTLS Authentication in Kyma
When the communication is two-way, like in the case of mTLS, both the client and the server are required to present signed certificates. Additionally, to establish a connection, both parties must trust the validity of the certificate presented by the other party. This means that the client must trust the CA that issued the server's certificate, and the server must trust the CA that issued the client's certificate.

When you develop applications in SAP BTP, Kyma runtime, Kyma acts as the server.

![mTLS Authentication](../assets/)

When a client attempts to connect to a server, the following steps take place:
//TODO fix the diagram and steps
1. The server presents its certificate, which contains its public key.

2. The client validates this certificate against its trusted certificates.

3. The client presents its own certificate, which also contains a public key.

4. The server validates the client’s certificate against its trusted certificates.

5. Both parties generate and exchange session keys using their respective public and private key pairs. 

After this, the authenticated bidirectional connection is established.

## Related Information
- [Configure mTLS Authentication for Your Workloads](./01-10-configure-mtls.md)