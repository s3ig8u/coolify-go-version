#!/bin/bash

echo "=== Terminal Diagnosis and Fix Script ==="
echo "Current shell: $SHELL"
echo "Current PATH: $PATH"
echo "Working directory: $(pwd)"
echo "User: $(whoami)"
echo "Date: $(date)"

echo ""
echo "=== Testing basic commands ==="
ls -la | head -5
echo "Git status:"
git status --porcelain
echo "Git branch:"
git branch

echo ""
echo "=== Attempting fixes ==="

# Fix 1: Reset shell environment
echo "1. Resetting shell environment..."
unset PROMPT_COMMAND
export PS1="$ "

# Fix 2: Check for hung processes
echo "2. Checking for hung processes..."
jobs

# Fix 3: Reset terminal settings
echo "3. Resetting terminal settings..."
reset 2>/dev/null || true
stty sane 2>/dev/null || true

# Fix 4: Source shell configuration
echo "4. Sourcing shell configuration..."
if [ -f ~/.zshrc ]; then
    source ~/.zshrc 2>/dev/null || true
fi
if [ -f ~/.bashrc ]; then
    source ~/.bashrc 2>/dev/null || true
fi

echo ""
echo "=== Testing after fixes ==="
echo "Current working directory: $(pwd)"
echo "Git status after fix:"
git status --short

echo ""
echo "=== Fix complete ==="
echo "Try running commands now..."
