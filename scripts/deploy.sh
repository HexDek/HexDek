#!/bin/bash
# HexDek deploy script — builds and deploys to DARKSTAR + MISTY
# Usage: ./scripts/deploy.sh [backend|frontend|both|frontend-dev]
#
# Targets:
#   backend       — cross-compile + ship hexdek-server to DARKSTAR (prod API)
#   frontend      — build + rsync React app to MISTY's ~/sites/hexdek/   (prod: hexdek.dev)
#   frontend-dev  — build + rsync React app to MISTY's ~/sites/hexdek-dev/ (staging: dev.hexdek.dev)
#   both          — backend + frontend (prod)
set -euo pipefail

DARKSTAR="josh@192.168.1.207"
MISTY="josh@192.168.1.200"
TARGET="${1:-both}"

deploy_backend() {
    echo "=== Building hexdek-server for Linux amd64 ==="
    GOOS=linux GOARCH=amd64 go build -o hexdek-server-linux ./cmd/hexdek-server/
    echo "Binary: $(ls -lh hexdek-server-linux | awk '{print $5}')"

    echo "=== Uploading to DARKSTAR ==="
    scp hexdek-server-linux "$DARKSTAR:/tmp/hexdek-server-new"

    echo "=== Swapping binary + restarting on port 8090 ==="
    ssh "$DARKSTAR" 'pkill -f hexdek-server || true; sleep 1; mv /tmp/hexdek-server-new $HOME/hexdek/hexdek-server && chmod +x $HOME/hexdek/hexdek-server && $HOME/hexdek/start-hexdek.sh'

    echo "=== Verifying ==="
    sleep 3
    ssh "$DARKSTAR" 'ss -tlnp | grep 8090 && tail -3 /tmp/hexdek-server.log'
    rm -f hexdek-server-linux
    echo "=== Backend deploy complete ==="
}

deploy_frontend() {
    echo "=== Building React frontend ==="
    cd hexdek && npm run build && cd ..

    echo "=== Deploying to MISTY (~/sites/hexdek/) ==="
    # CRITICAL: source is hexdek/dist/ (Vite React build), NOT web/ (old plain HTML MVP)
    # Target is ~/sites/hexdek/, NOT ~/hexdek/
    rsync -avz --delete hexdek/dist/ "$MISTY:~/sites/hexdek/"

    echo "=== Frontend deploy complete ==="
}

deploy_frontend_dev() {
    # Staging deploy — same build as prod, but lands in a separate directory
    # on MISTY served at dev.hexdek.dev. Both prod and dev point at the same
    # backend API on DARKSTAR (per the dev.hexdek.dev rollout plan), so this
    # is purely a visual-layer staging environment.
    local branch
    branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    echo "=== Building React frontend (branch: $branch) ==="
    cd hexdek && npm run build && cd ..

    echo "=== Deploying to MISTY (~/sites/hexdek-dev/) ==="
    rsync -avz --delete hexdek/dist/ "$MISTY:~/sites/hexdek-dev/"

    echo "=== Dev frontend deploy complete — live at dev.hexdek.dev ==="
}

case "$TARGET" in
    backend)      deploy_backend ;;
    frontend)     deploy_frontend ;;
    frontend-dev) deploy_frontend_dev ;;
    both)         deploy_backend; deploy_frontend ;;
    *)            echo "Usage: $0 [backend|frontend|both|frontend-dev]"; exit 1 ;;
esac
