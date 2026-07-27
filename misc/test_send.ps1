# --- CONFIGURATION ---
$smtpServer = "127.0.0.1"
$port = 2525
$emlPath = ".\test_email.eml"

if (-not (Test-Path $emlPath)) { Write-Error "File not found."; exit }

# 1. PARSE ENVELOPE INFO DIRECTLY FROM THE EML HEADERS
$fileLines = Get-Content $emlPath
$mailFrom = "test-sender@domain.com" # Fallback defaults
$rcptTo = "test-rcpt@yourdomain.com"

foreach ($line in $fileLines) {
    if ([string]::IsNullOrWhiteSpace($line)) { break }
    if ($line -match "^From:\s*(.*<)?(?<email>[^>]+)(>)?") { $mailFrom = $matches['email'].Trim() }
    if ($line -match "^To:\s*(.*<)?(?<email>[^>]+)(>)?") { $rcptTo = $matches['email'].Trim() }
}

Write-Host "Automated Handshake -> From: <$mailFrom> To: <$rcptTo>" -ForegroundColor Cyan

# 2. ESTABLISH NETWORK CONNECTION
$socket = New-Object System.Net.Sockets.TcpClient($smtpServer, $port)
$stream = $socket.GetStream()
$writer = New-Object System.IO.StreamWriter($stream)
$reader = New-Object System.IO.StreamReader($stream)

# Read complete SMTP response (handles multi-line responses like 250-...)
function Read-SmtpResponse {
    $response = ""
    do {
        $line = $reader.ReadLine()
        $response += $line + "`n"
    } while ($line -match '^\d{3}-') # Continue reading as long as line starts with 250-, 220-, etc.
    return $response
}

function Send-Command($command) {
    $writer.WriteLine($command)
    $writer.Flush()
    return Read-SmtpResponse
}

# 3. HANDSHAKE
$greeting = Read-SmtpResponse # Read initial 220 banner
$ehloResp = Send-Command "EHLO automated.tester.local"
$mailResp = Send-Command "MAIL FROM:<$mailFrom>"
$rcptResp = Send-Command "RCPT TO:<$rcptTo>"
$dataResp = Send-Command "DATA" # Should receive 354 Start mail input

# 4. STREAM RAW EML PAYLOAD
foreach ($line in $fileLines) {
    if ($line -eq ".") {
        $writer.WriteLine("..")
    } else {
        $writer.WriteLine($line)
    }
}
$writer.Flush()

# 5. CLOSE TRANSACTION
$endDataResp = Send-Command "."    # Waits for 250 OK
$quitResp    = Send-Command "QUIT" # Waits for 221 Bye

# Clean close
$writer.Close()
$stream.Close()
$socket.Close()

Write-Host "Success: EML automatically ingested!" -ForegroundColor Green