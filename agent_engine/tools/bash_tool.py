"""
2M Code — Bash Execution Tool

Executes shell commands with a 30-second timeout.
Commands ending with `&` run in background mode — they return immediately with the PID
instead of waiting for completion. This is useful for starting servers, watchers, etc.

Security: Commands run as the user's own process with no privilege escalation.
"""

import logging
import os
import signal
import subprocess
import time

logger = logging.getLogger("2mcode.tools.bash")

BASH_TOOL_DEFINITION = {
    "name": "bash",
    "description": "Execute a bash command. Returns stdout and stderr. "
                   "Timeout: 30 seconds. "
                   "If command ends with '&' it runs in background mode "
                   "and returns the PID immediately.",
    "input_schema": {
        "type": "object",
        "properties": {
            "command": {
                "type": "string",
                "description": "The bash command to run. "
                               "Append '&' at the end to run in background "
                               "(for starting servers, watchers, etc.).",
            }
        },
        "required": ["command"],
    },
}


def _execute_background(command: str) -> str:
    """
    Execute a command in background mode (non-blocking).

    Strips the trailing '&', starts the process detached from the parent,
    and returns immediately with the PID.

    Args:
        command: The shell command (should end with '&').

    Returns:
        String with the PID and confirmation message.
    """
    stripped = command.rstrip().rstrip("&").strip()

    logger.info("Starting background command (first 100 chars): %s", stripped[:100])

    try:
        proc = subprocess.Popen(
            stripped,
            shell=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            preexec_fn=lambda: os.setsid() if hasattr(os, "setsid") else None,
            start_new_session=True,
        )

        # Give it a moment to fail-fast (e.g. command not found)
        time.sleep(0.3)
        if proc.poll() is not None and proc.returncode != 0:
            return (
                f"Background command exited immediately with code {proc.returncode}. "
                f"Check if the command is valid: {stripped}"
            )

        return f"Background process started with PID {proc.pid}. Command: {stripped}"
    except FileNotFoundError:
        return f"Error: Command not found: {stripped.split()[0]}"
    except Exception as e:
        return f"Error starting background process: {e}"


def execute_bash(tool_input: dict) -> str:
    """
    Execute a bash command and return stdout + stderr.

    Args:
        tool_input: Dict with "command" key containing the shell command.

    Returns:
        Combined stdout and stderr output as a string.

    Raises:
        TimeoutError: If the command exceeds 30 seconds (blocking commands only).
    """
    command = tool_input.get("command", "")
    if not command:
        return "Error: No command provided."

    # Check for background mode (command ends with '&')
    if command.rstrip().endswith("&"):
        return _execute_background(command)

    logger.info("Executing bash command (first 100 chars): %s", command[:100])

    try:
        result = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            timeout=30,
        )
        output = result.stdout + result.stderr
        if result.returncode != 0:
            output += f"\n[exit code: {result.returncode}]"
        return output if output else "[no output]"
    except subprocess.TimeoutExpired:
        raise TimeoutError(f"Command timed out after 30 seconds: {command[:100]}")
