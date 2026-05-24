"""
2M Code — Plugin Base Class

Defines the Plugin interface with lifecycle hooks.
Users write a subclass of Plugin in a .py file under ~/.2mcode/plugins/
or .2mcode/plugins/ to extend the agent engine.
"""

import logging

logger = logging.getLogger("2mcode.plugin")


class Plugin:
    """Base class for all 2M Code plugins.

    Subclass this and override any hooks you need. All hooks are optional.

    Example:
        class MyPlugin(Plugin):
            name = "my_plugin"

            def on_agent_turn_start(self, req: dict) -> dict:
                req["system"] += "\\n[Plugin note: Be thorough!]"
                return req
    """

    name: str = "unnamed_plugin"

    def __init__(self):
        if self.name == "unnamed_plugin":
            self.name = type(self).__name__

    def on_startup(self, server_app):
        """Called once when the agent engine starts.

        Args:
            server_app: The FastAPI application instance.
        """

    def on_shutdown(self):
        """Called once when the agent engine shuts down."""

    def on_agent_turn_start(self, req: dict) -> dict:
        """Called before an agent turn.

        Return the (possibly modified) request dict, or the original.
        Args:
            req: The AgentRequest as a dict with keys:
                 provider, model, system, messages, tools, custom_tools, max_tokens, stream
        Returns:
            The (possibly modified) request dict.
        """
        return req

    def on_agent_turn_end(self, response: dict) -> dict:
        """Called after an agent turn completes.

        Return the (possibly modified) response dict, or the original.
        Args:
            response: The response dict with keys:
                      content, tool_calls, input_tokens, output_tokens
        Returns:
            The (possibly modified) response dict.
        """
        return response

    def on_tool_exec(self, tool_name: str, params: dict):
        """Called before a tool executes.

        Return a dict to override the tool result (skips actual execution).
        Return None to let the tool execute normally.
        Args:
            tool_name: Name of the tool being executed.
            params: Parameters passed to the tool.
        Returns:
            None to proceed, or a dict to override the result.
        """
        return None
