import JSEncrypt from "jsencrypt";

// encryptPassword encrypts a plaintext password using the given PEM-format RSA public key.
export function encryptPassword(
  publicKeyPEM: string,
  password: string
): string {
  const encrypt = new JSEncrypt();
  encrypt.setPublicKey(publicKeyPEM);
  const result = encrypt.encrypt(password);
  if (!result) {
    throw new Error("encryption failed");
  }
  return result;
}
