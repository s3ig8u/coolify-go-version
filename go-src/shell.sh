#!/bin/bash
# Simple wrapper to enter Nix development shell
source /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
nix develop --extra-experimental-features "nix-command flakes"
