if ($Host.Version.Major -lt 7) {throw "PowerShell 7 required!"}

# Replace with encrypted_caller_info from message you are interested in
$encrypted_caller_info = 'E4bNd2ERhDlUPFu7t4o5hHP/r1v6fDRl4jaFkLVW1DCU8Y2uQaHbXs7+XJY1AuaAYeGQbkEhjmpnJyYpUB+mvgOV6fsok97Zs8pZFddr8pnQ41iUpNPj3VNwWOQqBPmrwo/8nA=='

# Replace with METADATA_SECRET from bot config
$METADATA_SECRET = 'zRr5QP0ERaJq8VAQQ6N3wilI+GxQXoPymI1358qHBomJJA27UJQ2T2WxRi2NMANr'

### BELOW CODE
$encryptedBytes = [System.Convert]::FromBase64String($encrypted_caller_info)
$sha256 = [System.Security.Cryptography.SHA256]::Create()
$key = $sha256.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($METADATA_SECRET))
$nonceSize = 12 # AES-GCM uses nonce of 12 bytes
if ($encryptedBytes.Length -le $nonceSize) {
    throw "Encrypted data is too short"
}
$nonce = $encryptedBytes[0..($nonceSize-1)]
$cipherText = $encryptedBytes[$nonceSize..($encryptedBytes.Length-1)]
try {
    $aesGcm = [System.Security.Cryptography.AesGcm]::new($key)
    $tagSize = 16
    if ($cipherText.Length -le $tagSize) {
        throw "Ciphertext is too short to contain a tag"
    }
    $tag = $cipherText[($cipherText.Length-$tagSize)..($cipherText.Length-1)]
    $actualCipherText = $cipherText[0..($cipherText.Length-$tagSize-1)]
    $plaintextBytes = [byte[]]::new($actualCipherText.Length)
    $aesGcm.Decrypt(
            $nonce,
            $actualCipherText,
            $tag,
            $plaintextBytes,
            [byte[]]::new(0)  # Optional associated data (none in this case)
        )
    $plaintext = [System.Text.Encoding]::UTF8.GetString($plaintextBytes)
    $data = $plaintext | ConvertFrom-Json -ErrorAction Stop | ConvertTo-Json -Depth 5 -Compress:$false
    return $data
} finally {
    if ($null -ne $aesGcm) {
        $aesGcm.Dispose()
    }
}


