"""Example plugin: formats agent responses before returning to the user.

Demonstrates on_agent_turn_end hook by appending token usage info
and a timestamp to every agent response.
"""

import datetime

from plugin_base import Plugin


class ResponseFormatterPlugin(Plugin):
    name = "response_formatter"

    def on_agent_turn_end(self, response: dict) -> dict:
        content = response.get("content", "")
        tokens_in = response.get("input_tokens", 0)
        tokens_out = response.get("output_tokens", 0)

        footer = (
            f"\n\n---\n"
            f"*Response generated at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}*\n"
            f"*Tokens: {tokens_in} in / {tokens_out} out*"
        )
        response["content"] = content + footer
        return response
