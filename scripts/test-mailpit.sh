#!/usr/bin/env bash
#
# Send test emails to Mailpit via SMTP (port 1025).
# Usage: ./scripts/test-mailpit.sh [count]
#   count — number of test emails to send (default: 3)
#
# Mailpit must be running (docker compose up mailpit).
# View emails at http://localhost:8025

set -euo pipefail

SMTP_HOST="${SMTP_HOST:-localhost}"
SMTP_PORT="${SMTP_PORT:-1025}"
FROM="crm@bizmopol.local"
TO="client@example.com"
COUNT="${1:-3}"

command -v curl >/dev/null 2>&1 || { echo "curl is required but not installed."; exit 1; }

echo "Sending $COUNT test email(s) to Mailpit at $SMTP_HOST:$SMTP_PORT ..."
echo ""

for i in $(seq 1 "$COUNT"); do
    SUBJECT="Test email #$i - $(date '+%H:%M:%S')"
    MSG_ID="test-$(date +%s)-$i@bizmopol.local"
    BODY="This is test email number $i sent at $(date '+%Y-%m-%d %H:%M:%S').

It was sent by the BizmoPol CRM test script to verify SMTP delivery to Mailpit.

Regards,
BizmoPol CRM Test"

    curl --silent --show-error \
        --url "smtp://$SMTP_HOST:$SMTP_PORT" \
        --mail-from "$FROM" \
        --mail-rcpt "$TO" \
        --upload-file - <<EOF
From: BizmoPol CRM <$FROM>
To: Test Client <$TO>
Subject: $SUBJECT
Message-ID: <$MSG_ID>
Date: $(date -R 2>/dev/null || date '+%a, %d %b %Y %H:%M:%S %z')
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

$BODY
EOF

    echo "  [$i/$COUNT] Sent: $SUBJECT"
    [ "$i" -lt "$COUNT" ] && sleep 0.3
done

echo ""
echo "Done! Check Mailpit UI at http://localhost:8025"
