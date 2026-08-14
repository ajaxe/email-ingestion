# Email-Ingestion Infrastructure Details

## Build instructions for SMTP MTA at the edge

SMTP MTA for the edge will be build manually and deployed to remote linux VM. Use the following build commands:

_Assuming root of  the repo_

```pwsh
cd backend

$env:CGO_ENABLED=0; $env:GOOS="linux"; go build -trimpath -ldflags="-w -s" -o email-ingestion .
```
