#!/bin/bash

# Dev Container Post-Create Script
# Sets up the development environment with pnpm and Claude Code

# set -e  # Exit on any error

# Colors for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}==>${NC} ${1}"
}

print_success() {
    echo -e "${GREEN}✓${NC} ${1}"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} ${1}"
}

echo -e "${PURPLE}╔════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║        Dev Container Setup             ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════╝${NC}"
echo

# Setup pnpm
print_step "Setting up pnpm package manager..."
SHELL=$SHELL pnpm setup
print_success "pnpm setup completed"

# Reload environment to get pnpm in PATH
print_step "Reloading environment variables..."
eval "$(cat /home/vscode/.bashrc)"

# Verify pnpm installation
if [ -n "$PNPM_HOME" ]; then
    print_success "pnpm home directory configured: $PNPM_HOME"
else
    print_warning "pnpm home directory not found - this might be normal on first setup"
fi

# Setup git configuration from environment variables
print_step "Setting up git configuration..."
if [ -n "$GIT_EMAIL" ] && [ -n "$GIT_NAME" ]; then
    git config --global user.email "$GIT_EMAIL"
    git config --global user.name "$GIT_NAME"
    print_success "Git configured with email: $GIT_EMAIL and name: $GIT_NAME"
else
    print_warning "GIT_EMAIL and/or GIT_NAME environment variables not set - skipping git configuration"
fi

# Install Air for Go live reload
print_step "Installing Air live reload tool for Go..."
go install github.com/air-verse/air@v1.67.4
if command -v air &> /dev/null; then
    print_success "Air installed successfully at $(which air)"
else
    print_warning "Air installation could not be verified - check GOPATH/bin is in PATH"
fi

sudo chown -R $(whoami):$(whoami) /go/pkg

# Configure APP_BASE_URL so generated QR code links work from outside the
# container. Inside a Codespace, localhost:8080 isn't reachable by anything
# but this container, so point at the forwarded port's public hostname instead.
print_step "Configuring APP_BASE_URL..."
APP_ENV_FILE="$HOME/.app_env.sh"
if [ "$CODESPACES" = "true" ]; then
    echo "export APP_BASE_URL=\"https://${CODESPACE_NAME}-8080.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN}\"" > "$APP_ENV_FILE"
    print_success "APP_BASE_URL set to https://${CODESPACE_NAME}-8080.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN}"
    print_warning "Port 8080 must be set to Public visibility in the Ports tab for scanned QR codes to work off-machine"
else
    : > "$APP_ENV_FILE"
    print_warning "Not running in a Codespace - APP_BASE_URL will fall back to its default (http://localhost:8080)"
fi

for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [ -f "$rc" ] && ! grep -q "$APP_ENV_FILE" "$rc"; then
        printf '\n# App base URL (Codespaces port forwarding)\n[ -f "%s" ] && source "%s"\n' "$APP_ENV_FILE" "$APP_ENV_FILE" >> "$rc"
    fi
done

echo
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        Setup Complete! 🎉              ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo -e "${GREEN}Your development environment is ready.${NC}"
echo
