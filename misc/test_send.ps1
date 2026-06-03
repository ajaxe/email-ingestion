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
    # Stop parsing headers if we hit the blank line separating the body
    if ([string]::IsNullOrWhiteSpace($line)) { break }

    # Extract the email addresses using basic RegEx
    if ($line -match "^From:\s*(.*<)?(?<email>[^>]+)(>)?") {
        $mailFrom = $matches['email'].Trim()
    }
    if ($line -match "^To:\s*(.*<)?(?<email>[^>]+)(>)?") {
        $rcptTo = $matches['email'].Trim()
    }
}

Write-Host "Automated Handshake -> From: <$mailFrom> To: <$rcptTo>" -ForegroundColor Cyan

# 2. ESTABLISH NETWORK CONNECTION
$socket = New-Object System.Net.Sockets.TcpClient($smtpServer, $port)
$stream = $socket.GetStream()
$writer = New-Object System.IO.StreamWriter($stream)
$reader = New-Object System.IO.StreamReader($stream)

function Send-Command($command) {
    $writer.WriteLine($command)
    $writer.Flush()
    Start-Sleep -Milliseconds 100
    while ($stream.DataAvailable) { $null = $reader.ReadLine() }
}

# Clear initial greeting
Start-Sleep -Milliseconds 100
while ($stream.DataAvailable) { $null = $reader.ReadLine() }

# 3. DYNAMIC SMTP HANDSHAKE
Send-Command "EHLO automated.tester.local"
Send-Command "MAIL FROM:<$mailFrom>"
Send-Command "RCPT TO:<$rcptTo>"
Send-Command "DATA"

# 4. STREAM THE RAW EML PAYLOAD AS THE DATA BODY
foreach ($line in $fileLines) {
    if ($line -eq ".") {
        $writer.WriteLine("..") # Byte-stuffing safety
    } else {
        $writer.WriteLine($line)
    }
}
$writer.Flush()

# 5. CLOSE OUT TRANSACTION
Send-Command "."
Send-Command "QUIT"

$writer.Close(); $stream.Close(); $socket.Close()
Write-Host "Success: EML automatically ingested!" -ForegroundColor Green