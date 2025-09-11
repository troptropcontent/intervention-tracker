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
pnpm install -g @anthropic-ai/claude-code
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

sudo chown -R $(whoami):$(whoami) /go/pkg

echo
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        Setup Complete! 🎉              ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo -e "${GREEN}Your development environment is ready.${NC}"
echo
