#!/bin/bash

# Simple terminal test script
echo "Testing terminal output..." > terminal_test_output.txt
date >> terminal_test_output.txt
pwd >> terminal_test_output.txt
echo "If you can see this file, the terminal is working but output isn't returning to the AI assistant." >> terminal_test_output.txt

# Try to create the PR using GitHub CLI
echo "Attempting to create PR..." >> terminal_test_output.txt
gh pr create --title "feat: Add Go port of Coolify" --body "Complete Go port with distribution system" --base v4.x --head go >> terminal_test_output.txt 2>&1

echo "Check terminal_test_output.txt for results" >> terminal_test_output.txt
