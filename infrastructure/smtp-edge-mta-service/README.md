# Email Ingest Sytemd Service


## Remote Connection

EC2 instance has been setup with SSH key, we can login using it as follows:

```pwsh
ssh -i ~/.ssh/email_ingestion_mta_ec2 ec2-user@$(terraform output -raw public_ip)
```

## Uploading Artifacts (PowerShell)

To automatically query the public IP from Terraform and upload the required deployment artifacts (`output/email-ingest-arm64`, `setup.sh`, `uninstall.sh`, `email-ingest.service`, `config.yaml`) to the EC2 instance via SCP:

```pwsh
.\upload.ps1
```

## Quick Automated Setup

Run the setup script with `sudo` to perform user creation, directory permissions, journald retention setup, and service enablement in one command:

```bash
cd service_artifacts
chmod +x setup.sh
sudo ./setup.sh
```

## Quick Automated Uninstallation

Run the uninstall script with `sudo` to stop/disable the service, remove systemd unit & journald config, remove application files, and delete system user/group:

```bash
cd service_artifacts
chmod +x uninstall.sh
sudo ./uninstall.sh
```

## Manual Service Installation

Copy `email-ingest.service` to folder `/etc/systemd/system/` and run the following commands:

```bash
# Reload and start
sudo systemctl daemon-reload
sudo systemctl enable --now email-ingest
```

## Service Logs

To tail the logs:

```bash
sudo journalctl --namespace=email-ingest -u email-ingest -f
```

View logs since current boot:

```bash
sudo journalctl --namespace=email-ingest -u email-ingest -f
```

Filter by time window:

```bash
sudo journalctl --namespace=email-ingest -u email-ingest --since "1 hour ago"
```

Show only errors/warnings:

```bash
sudo journalctl --namespace=email-ingest -u email-ingest -p err..emerg
```

Manually truncate / vacuum only this service's logs:

```bash
# Truncate by size
sudo journalctl --namespace=email-ingest --vacuum-size=200M

# Truncate by age
sudo journalctl --namespace=email-ingest --vacuum-time=7d
```

## User Setup and Permissions

```bash
# 1. Create a dedicated system user
sudo useradd --system --no-create-home --shell /sbin/nologin email-ingest

# 2. Create the unified directory tree
sudo mkdir -p /opt/email-ingest/logs

# 3. Restrict permissions to the service user
sudo chown -R email-ingest:email-ingest /opt/email-ingest
sudo chmod 700 /opt/email-ingest
```

## Logs Setup

Using Journald to manage service specific logs. Configuration steps (one-time upon server setup):

1. **Create a Dedicated Journal Namespace**: `sudo mkdir -p /etc/systemd/journald@email-ingest.conf.d`
2. **Create Retention Config**: In directory `/etc/systemd/journald@email-ingest.conf.d`, create `retention.conf` with the following content:
```TOML
[Journal]
Storage=persistent
# Max disk space dedicated ONLY to email-ingest logs
SystemMaxUse=500M
# Keep at least this much disk space free on the volume
SystemKeepFree=1G
# Max size per rotated journal file
SystemMaxFileSize=50M
# Automatically drop service logs older than 14 days
MaxRetentionSec=14day
```