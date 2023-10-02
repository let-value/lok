# Lok

Lok is a CLI tool designed for encrypting files in place using only a password and a glob pattern.

## Installation

Installation instructions will be provided soon.

## Usage

### Encrypting Files

```bash
lok encrypt [glob pattern] [password]
```

### Decrypting Files

```bash
lok decrypt [glob pattern] [password]
```

For more advanced usage and options, refer to the source code or run `lok --help`.

## Security

### Encryption Approach

Key Derivation: Lok uses the Argon2id variant of the Argon2 password hashing function to derive a cryptographic key from the user's password. This is combined with a static "pepper" to further enhance the security of the derived key.

### Encryption Algorithm

Lok uses the ChaCha20-Poly1305 authenticated encryption algorithm.

### Nonce Generation

There are two methods of nonce generation in the code:

- Weak Method: This method utilizes a predetermined nonce and is specifically used for encrypting directory and file names. Relying on a fixed nonce with a consistent key poses a significant security risk, as it can undermine the confidentiality of the encrypted data.
  
- Secure Method: This method derives the nonce from the data set to be encrypted, using the SHA-256 hashing algorithm. It's primarily used for encrypting file contents, leveraging the original file name to derive the nonce. Although this approach is an improvement over the weak method, it doesn't align with best practices. Ideally, nonces should be both random and unique for every encryption task.

## Disclaimer

While Lok aims to provide robust encryption, users should always backup their data before using any encryption tool. The developers are not responsible for any data loss or damages. Additionally, users should be cautious about the nonce generation methods and avoid using the weak encryption method.

## Contributing

We welcome contributions from the community. Whether you're fixing bugs, improving the documentation, or proposing new features, your efforts are appreciated!

## License

[MIT](LICENSE)