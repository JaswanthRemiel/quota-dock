using System;
using System.IO;
using System.Security.Cryptography;
using System.Text;

namespace QuotaDock.Windows.Services
{
    /// <summary>
    /// Implements zero-plaintext Windows DPAPI (Data Protection API) credential storage for QuotaDock.
    /// Equivalent to macOS Keychain Security.framework on Apple platforms.
    /// </summary>
    public static class WindowsCredentialManager
    {
        private static readonly string CredentialStorePath = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
            "QuotaDock",
            "credentials.bin"
        );

        private static readonly byte[] Entropy = Encoding.UTF8.GetBytes("QuotaDock_Windows_DPAPI_Entropy");

        public static void SaveSecret(string keyName, string secret)
        {
            var dir = Path.GetDirectoryName(CredentialStorePath);
            if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
            {
                Directory.CreateDirectory(dir);
            }

            var plainBytes = Encoding.UTF8.GetBytes($"{keyName}={secret}");
            var cipherBytes = ProtectedData.Protect(
                plainBytes,
                Entropy,
                DataProtectionScope.CurrentUser
            );

            File.WriteAllBytes(CredentialStorePath, cipherBytes);
        }

        public static string? LoadSecret(string keyName)
        {
            if (!File.Exists(CredentialStorePath))
            {
                return null;
            }

            try
            {
                var cipherBytes = File.ReadAllBytes(CredentialStorePath);
                var plainBytes = ProtectedData.Unprotect(
                    cipherBytes,
                    Entropy,
                    DataProtectionScope.CurrentUser
                );

                var decoded = Encoding.UTF8.GetString(plainBytes);
                var parts = decoded.Split('=', 2);
                if (parts.Length == 2 && parts[0] == keyName)
                {
                    return parts[1];
                }
            }
            catch
            {
                // Decryption failure or corrupted store
            }

            return null;
        }

        public static void Clear()
        {
            if (File.Exists(CredentialStorePath))
            {
                File.Delete(CredentialStorePath);
            }
        }
    }
}
