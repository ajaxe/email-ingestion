#!/usr/bin/env bash
# ==============================================================================
# Setup Script for Email Ingest Edge MTA Service on Amazon Linux
# ==============================================================================
# This script performs host initialization for the email-ingest service:
# 1. System user & group creation, directory permissions (/opt/email-ingest).
# 2. Deploying service binary (email-ingest-arm64 -> /opt/email-ingest/email-ingest) & config.yaml.
# 3. Journald namespace logging & log retention configuration (idempotent).
# 4. Systemd service unit installation (/etc/systemd/system/email-ingest.service).
# 5. Reloading systemd, enabling, and starting the email-ingest service safely.
# ==============================================================================

set -euo pipefail

# Ensure script is executed with root privileges
if [[ $EUID -ne 0 ]]; then
    echo "[ERROR] This script must be run as root or via sudo." >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_SRC="${SCRIPT_DIR}/email-ingest.service"
if [[ ! -f "${SERVICE_SRC}" && -f "${SCRIPT_DIR}/service_artifacts/email-ingest.service" ]]; then
    SERVICE_SRC="${SCRIPT_DIR}/service_artifacts/email-ingest.service"
fi
SERVICE_DEST="/etc/systemd/system/email-ingest.service"
JOURNAL_CONF_DIR="/etc/systemd/journald@email-ingest.conf.d"
JOURNAL_CONF_FILE="${JOURNAL_CONF_DIR}/retention.conf"
APP_DIR="/opt/email-ingest"
APP_BIN="${APP_DIR}/email-ingest"
APP_CONFIG="${APP_DIR}/config.yaml"

echo "=========================================================="
echo " Starting Email Ingest Service Setup"
echo "=========================================================="

# Check if systemd is running as PID 1
IS_SYSTEMD_ACTIVE=0
if [[ -d /run/systemd/system ]] && command -v systemctl &>/dev/null; then
    IS_SYSTEMD_ACTIVE=1
fi

# ------------------------------------------------------------------------------
# 1. User Setup and Permissions
# ------------------------------------------------------------------------------
echo "[1/5] Configuring system user and directories..."

if getent group email-ingest &>/dev/null; then
    echo "  - System group 'email-ingest' already exists."
else
    echo "  - Creating system group 'email-ingest'..."
    groupadd --system email-ingest
fi

if getent passwd email-ingest &>/dev/null; then
    echo "  - System user 'email-ingest' already exists."
else
    echo "  - Creating system user 'email-ingest'..."
    useradd --system --no-create-home --shell /sbin/nologin -g email-ingest email-ingest
fi

echo "  - Setting up directory tree ${APP_DIR}/logs..."
mkdir -p "${APP_DIR}/logs"

echo "  - Restricting ownership and permissions to 'email-ingest'..."
chown -R email-ingest:email-ingest "${APP_DIR}"
chmod 700 "${APP_DIR}"
chmod 700 "${APP_DIR}/logs"

# ------------------------------------------------------------------------------
# 2. Deploying Binary and Configuration File
# ------------------------------------------------------------------------------
echo "[2/5] Deploying service binary and configuration file..."

SRC_BIN=""
if [[ -f "${SCRIPT_DIR}/email-ingest-arm64" ]]; then
    SRC_BIN="${SCRIPT_DIR}/email-ingest-arm64"
elif [[ -f "${SCRIPT_DIR}/email-ingest" ]]; then
    SRC_BIN="${SCRIPT_DIR}/email-ingest"
elif [[ -f "${SCRIPT_DIR}/service_artifacts/email-ingest-arm64" ]]; then
    SRC_BIN="${SCRIPT_DIR}/service_artifacts/email-ingest-arm64"
elif [[ -f "${SCRIPT_DIR}/service_artifacts/email-ingest" ]]; then
    SRC_BIN="${SCRIPT_DIR}/service_artifacts/email-ingest"
fi

if [[ -n "${SRC_BIN}" ]]; then
    echo "  - Copying binary from ${SRC_BIN} to ${APP_BIN}..."
    cp "${SRC_BIN}" "${APP_BIN}"
    chmod 755 "${APP_BIN}"
    chown email-ingest:email-ingest "${APP_BIN}"
    echo "  - Binary installed successfully."
elif [[ -f "${APP_BIN}" ]]; then
    echo "  - Source binary not found in ${SCRIPT_DIR} or ${SCRIPT_DIR}/service_artifacts; retaining existing ${APP_BIN}."
else
    echo "  - [WARNING] No binary found at ${SCRIPT_DIR}/email-ingest-arm64, ${SCRIPT_DIR}/service_artifacts/email-ingest-arm64, or ${APP_BIN}."
fi

SRC_CONFIG="${SCRIPT_DIR}/config.yaml"
if [[ ! -f "${SRC_CONFIG}" && -f "${SCRIPT_DIR}/service_artifacts/config.yaml" ]]; then
    SRC_CONFIG="${SCRIPT_DIR}/service_artifacts/config.yaml"
fi

if [[ -f "${SRC_CONFIG}" ]]; then
    echo "  - Copying configuration file from ${SRC_CONFIG} to ${APP_CONFIG}..."
    cp "${SRC_CONFIG}" "${APP_CONFIG}"
    chmod 640 "${APP_CONFIG}"
    chown email-ingest:email-ingest "${APP_CONFIG}"
    echo "  - Configuration file installed successfully."
elif [[ -f "${APP_CONFIG}" ]]; then
    echo "  - Source configuration file not found in ${SCRIPT_DIR} or ${SCRIPT_DIR}/service_artifacts; retaining existing ${APP_CONFIG}."
else
    echo "  - [WARNING] No configuration file found at ${SRC_CONFIG} or ${APP_CONFIG}."
fi

# ------------------------------------------------------------------------------
# 3. Journald Namespace Logging Setup
# ------------------------------------------------------------------------------
echo "[3/5] Setting up systemd journald namespace retention rules..."

mkdir -p "${JOURNAL_CONF_DIR}"

TMP_JOURNAL_CONF="$(mktemp)"
trap 'rm -f "${TMP_JOURNAL_CONF}"' EXIT

cat << 'EOF' > "${TMP_JOURNAL_CONF}"
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
EOF

JOURNAL_CONFIG_CHANGED=0
if [[ ! -f "${JOURNAL_CONF_FILE}" ]] || ! cmp -s "${TMP_JOURNAL_CONF}" "${JOURNAL_CONF_FILE}"; then
    cp "${TMP_JOURNAL_CONF}" "${JOURNAL_CONF_FILE}"
    chmod 644 "${JOURNAL_CONF_FILE}"
    JOURNAL_CONFIG_CHANGED=1
    echo "  - Configured ${JOURNAL_CONF_FILE}"
else
    echo "  - Configuration ${JOURNAL_CONF_FILE} is unchanged."
fi
rm -f "${TMP_JOURNAL_CONF}"

if [[ ${JOURNAL_CONFIG_CHANGED} -eq 1 ]]; then
    if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
        echo "  - Restarting systemd-journald to activate updated configuration..."
        systemctl restart systemd-journald
    else
        echo "  - [NOTICE] Systemd not active; skipping systemd-journald restart."
    fi
else
    echo "  - Skipping systemd-journald restart (configuration unchanged)."
fi

# ------------------------------------------------------------------------------
# 4. Installing Systemd Service
# ------------------------------------------------------------------------------
echo "[4/5] Installing systemd service unit..."

if [[ ! -f "${SERVICE_SRC}" ]]; then
    echo "[ERROR] Source service file not found at ${SERVICE_SRC}" >&2
    exit 1
fi

SERVICE_CHANGED=0
if [[ ! -f "${SERVICE_DEST}" ]] || ! cmp -s "${SERVICE_SRC}" "${SERVICE_DEST}"; then
    cp "${SERVICE_SRC}" "${SERVICE_DEST}"
    chmod 644 "${SERVICE_DEST}"
    SERVICE_CHANGED=1
    echo "  - Installed updated ${SERVICE_DEST}"
else
    echo "  - Service unit ${SERVICE_DEST} is unchanged."
fi

if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
    if [[ ${SERVICE_CHANGED} -eq 1 ]]; then
        echo "  - Reloading systemd manager configuration..."
        systemctl daemon-reload
    else
        echo "  - Skipping daemon-reload (service unit unchanged)."
    fi
else
    echo "  - [NOTICE] Systemd not active; skipping daemon-reload."
fi

# ------------------------------------------------------------------------------
# 5. Service Activation
# ------------------------------------------------------------------------------
echo "[5/5] Enabling and starting email-ingest service..."

if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
    echo "  - Enabling email-ingest service..."
    systemctl enable email-ingest

    if [[ -f "${APP_BIN}" ]]; then
        echo "  - Starting/restarting email-ingest service..."
        if systemctl is-active --quiet email-ingest; then
            systemctl restart email-ingest
        else
            systemctl start email-ingest
        fi
    else
        echo "  - [NOTICE] Service binary not found at ${APP_BIN}."
        echo "             Service unit enabled. It will start automatically once binary is deployed."
    fi
else
    echo "  - [NOTICE] Systemd not active; skipping service enable/start."
fi

echo "=========================================================="
echo " Setup complete!"
if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
    echo " Service Status: $(systemctl is-active email-ingest || true)"
    echo " To inspect service status: sudo systemctl status email-ingest"
    echo " To tail service logs:     sudo journalctl -u email-ingest -f"
else
    echo " Systemd environment is not running."
fi
echo "=========================================================="
