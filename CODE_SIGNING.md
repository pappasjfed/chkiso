# Code Signing Setup

To enable code signing for the Windows executable, you need to set up the following secrets and variables in your GitHub repository settings.

## Required Secrets

Go to **Settings** → **Secrets and variables** → **Actions** and add:

1. **CERTIFICATE** (Secret)
   - The base64-encoded PFX certificate
   - To create in PowerShell:
     ```powershell
     $bytes = [System.IO.File]::ReadAllBytes("certificate.pfx")
     $base64 = [System.Convert]::ToBase64String($bytes)
     $base64 | Out-File -FilePath "certificate-base64.txt"
     ```
   - Or using certutil:
     ```cmd
     certutil -encode certificate.pfx certificate-base64.txt
     ```
   - Then remove the `-----BEGIN CERTIFICATE-----` and `-----END CERTIFICATE-----` lines and any whitespace/newlines to create a single base64 string

2. **CERT_PASSWORD** (Secret)
   - The password for the PFX certificate

## Required Variables

Go to **Settings** → **Secrets and variables** → **Actions** → **Variables** tab and add:

1. **CERTHASH** (Variable)
   - The SHA1 thumbprint of your certificate
   - To find: Double-click your certificate in Windows, go to Details tab, scroll to Thumbprint
   - Example: `1234567890ABCDEF1234567890ABCDEF12345678`

2. **CERTNAME** (Variable)
   - The subject name of the certificate
   - To find: Open the certificate, look at "Issued to" field
   - Example: `YourCompanyName` or `CN=YourCompanyName`

## Obtaining a Code Signing Certificate

You can obtain a code signing certificate from:
- Certificate Authorities (DigiCert, Sectigo, GlobalSign, etc.)
- Self-signed certificate for testing (not recommended for production)

### Creating a Self-Signed Certificate (Testing Only)

For testing purposes only, you can create a self-signed certificate:

```powershell
# Create a self-signed certificate
$cert = New-SelfSignedCertificate `
    -Type CodeSigningCert `
    -Subject "CN=Test Code Signing" `
    -KeyAlgorithm RSA `
    -KeyLength 2048 `
    -Provider "Microsoft Enhanced RSA and AES Cryptographic Provider" `
    -CertStoreLocation "Cert:\CurrentUser\My" `
    -NotAfter (Get-Date).AddYears(2)

# Export to PFX
$password = ConvertTo-SecureString -String "YourPassword" -Force -AsPlainText
Export-PfxCertificate -Cert $cert -FilePath "test-certificate.pfx" -Password $password

# Get thumbprint
$cert.Thumbprint
```

**Note**: Self-signed certificates will still show warnings in Windows. For production, you need a certificate from a trusted CA.

## Verification

Once configured, the workflows will automatically sign the executable during builds. The signing step will be skipped if the secrets/variables are not set, allowing the workflow to continue without signing.

To verify signing worked:
1. Download the executable from the artifacts
2. Right-click the file → Properties → Digital Signatures tab
3. You should see the signature information
