function base64ToArrayBuffer(base64: string): ArrayBuffer {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

// encryptPassword encrypts a plaintext password using the given base64-encoded RSA public key (SPKI format).
export async function encryptPassword(
  publicKeyBase64: string,
  password: string
): Promise<string> {
  const der = base64ToArrayBuffer(publicKeyBase64);
  const publicKey = await crypto.subtle.importKey(
    "spki",
    der,
    { name: "RSA-OAEP", hash: "SHA-256" },
    false,
    ["encrypt"]
  );
  const encoded = new TextEncoder().encode(password);
  const encrypted = await crypto.subtle.encrypt(
    { name: "RSA-OAEP" },
    publicKey,
    encoded
  );
  return arrayBufferToBase64(encrypted);
}
