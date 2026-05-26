"""Example plugin: filters dangerous arguments from tool calls.

Intercepts tool execution to block potentially destructive operations
like recursive deletion or command injection patterns.
"""

from plugin_base import Plugin


class SafetyFilterPlugin(Plugin):
    name = "safety_filter"

    DANGEROUS_PATTERNS = [
        "rm -rf /",
        "rm -rf ~",
        "rm -rf .",
        "format(",
        "drop database",
        "DROP DATABASE",
    ]

    def on_tool_exec(self, tool_name: str, params: dict) -> dict | None:
        params_str = str(params)
        for pattern in self.DANGEROUS_PATTERNS:
            if pattern in params_str:
                return {
                    "success": False,
                    "error": f"Blocked by safety_filter plugin: potentially dangerous pattern '{pattern}' detected",
                    "output": "",
                }
        return None
