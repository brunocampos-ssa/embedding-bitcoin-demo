# External signer

[English](../en-US/16-external-signer.md) | [Português do Brasil](../pt-BR/16-external-signer.md)

The SDK exposes `ConnectWithSigner`, builder variants, and signer interfaces. Production should evaluate HSM/MPC-backed signing so the API host does not hold a plaintext mnemonic. Add approval policy, availability planning, key rotation, and tested recovery before deployment.
