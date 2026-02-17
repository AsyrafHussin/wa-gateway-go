#!/bin/sh
set -e

ENV_FILE=".env"

if [ -f "$ENV_FILE" ]; then
    printf "\n.env already exists. Overwrite? [y/N] "
    read -r answer
    case "$answer" in
        [yY]*) ;;
        *) echo "Aborted."; exit 0 ;;
    esac
fi

generate_key() {
    # 32-byte random hex string
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 32
    elif [ -r /dev/urandom ]; then
        head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'
    else
        date +%s%N | sha256sum | head -c 64
    fi
}

echo ""
echo "wa-gateway-go setup"
echo "==================="
echo ""
echo "  1) Auto-generate (recommended — sets random API_KEY, sensible defaults)"
echo "  2) Manual (prompts for key values)"
echo ""
printf "Choose [1/2]: "
read -r mode

case "$mode" in
    2)
        printf "\nAPI_KEY (required): "
        read -r api_key
        if [ -z "$api_key" ]; then
            echo "API_KEY cannot be empty."
            exit 1
        fi

        printf "PORT [4010]: "
        read -r port
        port="${port:-4010}"

        printf "WEBHOOK_URL (optional, press Enter to skip): "
        read -r webhook_url

        printf "WEBHOOK_SECRET (optional, press Enter to skip): "
        read -r webhook_secret

        printf "WS_ALLOWED_ORIGINS [*]: "
        read -r ws_origins
        ws_origins="${ws_origins:-*}"
        ;;
    *)
        api_key=$(generate_key)
        port="4010"
        webhook_url=""
        webhook_secret=""
        ws_origins="*"
        ;;
esac

cat > "$ENV_FILE" <<EOF
# Server
PORT=${port}
HOST=0.0.0.0
API_KEY=${api_key}
LOG_LEVEL=info

# CORS
CORS_ORIGINS=*

# Phone validation
PHONE_COUNTRY_CODE=60
PHONE_MIN_LENGTH=11
PHONE_MAX_LENGTH=12

# WhatsApp
DATA_DIR=./data
TYPING_DELAY_MS=1000
AUTO_READ_RECEIPT=false

# Webhook
WEBHOOK_URL=${webhook_url}
WEBHOOK_SECRET=${webhook_secret}
WEBHOOK_TIMEOUT_MS=5000

# Rate limits (requests per minute)
RATE_LIMIT_DEVICES=10
RATE_LIMIT_MESSAGES=30
RATE_LIMIT_VALIDATE=60

# Cache
CACHE_TTL_SECONDS=3600

# WebSocket
WS_ALLOWED_ORIGINS=${ws_origins}
WS_AUTH_TIMEOUT=5
EOF

echo ""
echo ".env created successfully!"
echo ""
echo "  API_KEY: ${api_key}"
echo "  PORT:    ${port}"
echo ""
echo "Run:  make docker    (Docker)"
echo "      make run       (local)"
