#!/usr/bin/env bash
# ==============================================================================
# Uninstall Script for Email Ingest Edge MTA Service on Amazon Linux
# ==============================================================================
# This script performs host clean up for the email-ingest service:
# 1. Stopping and disabling systemd service (email-ingest).
# 2. Removing systemd service unit (/etc/systemd/system/email-ingest.service).
# 3. Removing journald retention configuration (/etc/systemd/journald@email-ingest.conf.d).
# 4. Removing application directory (/opt/email-ingest).
# 5. Removing system user & group (email-ingest).
# ==============================================================================

set -euo pipefail

# Ensure script is executed with root privileges
if [[ $EUID -ne 0 ]]; then
    echo "[ERROR] This script must be run as root or via sudo." >&2
    exit 1
fi

SERVICE_DEST="/etc/systemd/system/email-ingest.service"
JOURNAL_CONF_DIR="/etc/systemd/journald@email-ingest.conf.d"
JOURNAL_CONF_FILE="${JOURNAL_CONF_DIR}/retention.conf"
APP_DIR="/opt/email-ingest"

echo "=========================================================="
echo " Starting Email Ingest Service Uninstallation"
echo "=========================================================="

# Check if systemd is running as PID 1
IS_SYSTEMD_ACTIVE=0
if [[ -d /run/systemd/system ]] && command -v systemctl &>/dev/null; then
    IS_SYSTEMD_ACTIVE=1
fi

# ------------------------------------------------------------------------------
# 1. Stopping and Disabling Service
# ------------------------------------------------------------------------------
echo "[1/5] Stopping and disabling email-ingest service..."

if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
    if systemctl is-active --quiet email-ingest 2>/dev/null; then
        echo "  - Stopping email-ingest service..."
        systemctl stop email-ingest
    else
        echo "  - Service email-ingest is not active."
    fi

    if systemctl is-enabled --quiet email-ingest 2>/dev/null; then
        echo "  - Disabling email-ingest service..."
        systemctl disable email-ingest
    else
        echo "  - Service email-ingest is not enabled."
    fi
else
    echo "  - [NOTICE] Systemd not active; skipping service stop/disable."
fi

# ------------------------------------------------------------------------------
# 2. Removing Systemd Service Unit
# ------------------------------------------------------------------------------
echo "[2/5] Removing systemd service unit..."

if [[ -f "${SERVICE_DEST}" ]]; then
    echo "  - Removing ${SERVICE_DEST}..."
    rm -f "${SERVICE_DEST}"
    if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
        echo "  - Reloading systemd manager configuration..."
        systemctl daemon-reload
        systemctl reset-failed email-ingest 2>/dev/null || true
    fi
else
    echo "  - Service file ${SERVICE_DEST} does not exist."
fi

# ------------------------------------------------------------------------------
# 3. Removing Journald Retention Configuration
# ------------------------------------------------------------------------------
echo "[3/5] Removing journald namespace retention rules..."

JOURNAL_CONFIG_REMOVED=0
if [[ -f "${JOURNAL_CONF_FILE}" ]]; then
    echo "  - Removing configuration ${JOURNAL_CONF_FILE}..."
    rm -f "${JOURNAL_CONF_FILE}"
    JOURNAL_CONFIG_REMOVED=1
fi

if [[ -d "${JOURNAL_CONF_DIR}" ]]; then
    echo "  - Removing journal configuration directory ${JOURNAL_CONF_DIR}..."
    rm -rf "${JOURNAL_CONF_DIR}"
    JOURNAL_CONFIG_REMOVED=1
fi

if [[ ${JOURNAL_CONFIG_REMOVED} -eq 1 ]]; then
    if [[ ${IS_SYSTEMD_ACTIVE} -eq 1 ]]; then
        echo "  - Restarting systemd-journald to activate updated configuration..."
        systemctl restart systemd-journald
    else
        echo "  - [NOTICE] Systemd not active; skipping systemd-journald restart."
    fi
else
    echo "  - Journald configuration files not present."
fi

# ------------------------------------------------------------------------------
# 4. Removing Application Directory
# ------------------------------------------------------------------------------
echo "[4/5] Removing application directory..."

if [[ -d "${APP_DIR}" ]]; then
    echo "  - Removing ${APP_DIR}..."
    rm -rf "${APP_DIR}"
    echo "  - Application directory removed successfully."
else
    echo "  - Application directory ${APP_DIR} does not exist."
fi

# ------------------------------------------------------------------------------
# 5. Removing System User and Group
# ------------------------------------------------------------------------------
echo "[5/5] Removing system user and group..."

if getent passwd email-ingest &>/dev/null; then
    echo "  - Removing system user 'email-ingest'..."
    userdel email-ingest
else
    echo "  - System user 'email-ingest' does not exist."
fi

if getent group email-ingest &>/dev/null; then
    echo "  - Removing system group 'email-ingest'..."
    groupdel email-ingest
else
    echo "  - System group 'email-ingest' does not exist."
fi

echo "=========================================================="
echo " Uninstallation complete!"
echo " All changes made by setup.sh have been reverted."
echo "=========================================================="
