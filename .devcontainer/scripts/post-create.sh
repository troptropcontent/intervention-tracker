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

# Install Claude Code globally
print_step "Installing Claude Code CLI globally..."
curl -fsSL https://claude.ai/install.sh | bash
print_success "Claude Code CLI installed successfully"

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
go install github.com/air-verse/air@latest
if command -v air &> /dev/null; then
    print_success "Air installed successfully at $(which air)"
else
    print_warning "Air installation could not be verified - check GOPATH/bin is in PATH"
fi

sudo chown -R $(whoami):$(whoami) /go/pkg

# Auto-source resolved 1Password secrets (~/.env.local.sh) in every new shell,
# so commands never need to be prefixed with `op run --env-file=./.env --`.
# The file itself is (re)generated on every container start by
# .devcontainer/scripts/resolve-env.sh (postStartCommand).
print_step "Wiring up auto-loading of resolved secrets..."
SOURCE_SNIPPET='[ -f "$HOME/.env.local.sh" ] && source "$HOME/.env.local.sh"'
for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [ -f "$rc" ] && ! grep -qF ".env.local.sh" "$rc"; then
        {
            echo ""
            echo "# Load 1Password-resolved secrets (see .devcontainer/scripts/resolve-env.sh)"
            echo "$SOURCE_SNIPPET"
        } >> "$rc"
        print_success "Added secret auto-load to $rc"
    fi
done

echo
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        Setup Complete! 🎉              ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo -e "${GREEN}Your development environment is ready.${NC}"
echo
