#!/bin/sh
# NeetoDesk CLI — session-start hook for Claude Code
# Lightweight auth liveness check. Always exits 0 (informational).

if ! command -v neetodesk >/dev/null 2>&1; then
  echo "NeetoDesk CLI is not installed or not on PATH."
  exit 0
fi

if neetodesk whoami >/dev/null 2>&1; then
  echo "NeetoDesk plugin active."
else
  echo "NeetoDesk CLI installed but not authenticated. Run 'neetodesk login' to authenticate."
fi

exit 0
