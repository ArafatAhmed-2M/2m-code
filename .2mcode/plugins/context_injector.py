"""Example plugin: injects coding guidelines into every agent's system prompt.

Place this in ~/.2mcode/plugins/ or .2mcode/plugins/ to activate.
"""

from plugin_base import Plugin


class ContextInjectorPlugin(Plugin):
    name = "context_injector"

    GUIDELINES = """
## Project Guidelines (injected by context_injector plugin)
- Prefer simple, readable code over clever optimizations
- Always handle errors gracefully with meaningful messages
- Use descriptive variable names
- Include a brief comment for complex logic
"""

    def on_agent_turn_start(self, req: dict) -> dict:
        system = req.get("system", "")
        req["system"] = system + self.GUIDELINES
        return req
