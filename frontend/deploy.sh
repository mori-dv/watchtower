#!/usr/bin/env bash
# ==============================================================================
# 🚀 1-Click Automated VPS Update Engine (SSH Reverse Tunnel)
# Target: https://watchtower.omnicraft.ir
# ==============================================================================

set -euo pipefail

# Text color formatting
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}${BOLD}   Watchtower // 1-Click Fast Update & Deploy Engine     ${NC}"
echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 1. Load configuration from .deploy.env if exists
if [ -f ".deploy.env" ]; then
    # shellcheck disable=SC1091
    source ".deploy.env"
fi

VPS_HOST="${VPS_HOST:-31.57.47.10}"
VPS_USER="${VPS_USER:-mori}"
VPS_PORT="${VPS_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/var/www/watchtower}"
DOMAIN="${DOMAIN:-watchtower.omnicraft.ir}"
TUNNEL_PORT="${TUNNEL_PORT:-8899}"

echo -e "📡 ${BOLD}Target Server:${NC} ${VPS_USER}@${VPS_HOST}:${VPS_PORT}"
echo -e "📁 ${BOLD}Destination:${NC}   ${REMOTE_DIR}"
echo -e "🌐 ${BOLD}Domain:${NC}        https://${DOMAIN}"
echo ""

# 2. Local Production Build
echo -e "${CYAN}📦 [1/3] Building production assets (Vite)...${NC}"
npm run build || node ./node_modules/vite/bin/vite.js build
if [ ! -d "dist" ] || [ ! -f "dist/index.html" ]; then
    echo -e "${RED}❌ Build failed: dist/index.html not found!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build succeeded.${NC}"

# 3. Create update zip package
echo -e "${CYAN}📦 [2/3] Compressing update bundle...${NC}"
LOCAL_ZIP="/tmp/watchtower_update.zip"
rm -f "$LOCAL_ZIP" 2>/dev/null || true
(cd dist && zip -rq -9 "$LOCAL_ZIP" .)

ZIP_SIZE=$(du -h "$LOCAL_ZIP" | awk '{print $1}')
echo -e "${GREEN}✓ Update bundle created (${ZIP_SIZE}).${NC}"

# 4. Start local tunnel HTTP server in background
echo -e "${CYAN}⚡ [3/3] Starting internal SSH tunnel transfer...${NC}"
python3 -m http.server "$TUNNEL_PORT" --directory /tmp --bind 127.0.0.1 >/dev/null 2>&1 &
PY_PID=$!

cleanup() {
    kill "$PY_PID" 2>/dev/null || true
    rm -f "$LOCAL_ZIP" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Allow local server a split second to bind
sleep 0.5

# 5. Connect via SSH with Reverse Tunnel (-R) and update files on VPS
ssh -R "${TUNNEL_PORT}:localhost:${TUNNEL_PORT}" -t -p "$VPS_PORT" "$VPS_USER@$VPS_HOST" "
    set -e
    echo '📥 Downloading update through encrypted SSH tunnel...'
    curl -sS -o /tmp/watchtower_update.zip http://localhost:${TUNNEL_PORT}/watchtower_update.zip

    echo '📂 Unpacking new assets to ${REMOTE_DIR}...'
    sudo mkdir -p /tmp/wt_new
    sudo unzip -qo /tmp/watchtower_update.zip -d /tmp/wt_new
    sudo mkdir -p ${REMOTE_DIR}
    sudo cp -rf /tmp/wt_new/* ${REMOTE_DIR}/
    sudo rm -rf /tmp/wt_new /tmp/watchtower_update.zip

    echo '🔒 Updating permissions...'
    sudo chown -R www-data:www-data ${REMOTE_DIR}
    sudo chmod -R 755 ${REMOTE_DIR}

    echo '🔄 Reloading Nginx...'
    sudo nginx -t && sudo systemctl reload nginx
    echo '✓ Nginx reloaded successfully.'
"

echo ""
echo -e "${GREEN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}${BOLD}   🎉 UPDATE DEPLOYED SUCCESSFULLY!                      ${NC}"
echo -e "${GREEN}${BOLD}   Website: https://${DOMAIN}                           ${NC}"
echo -e "${GREEN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
