"""Example plugin: logs every agent turn to a file."""

import datetime
import os

from plugin_base import Plugin


class TurnLoggerPlugin(Plugin):
    name = "turn_logger"

    def __init__(self):
        super().__init__()
        self.log_file = os.path.expanduser("~/.2mcode/plugin_turn_log.txt")

    def on_agent_turn_start(self, req: dict) -> dict:
        ts = datetime.datetime.now().isoformat()
        model = req.get("model", "?")
        system_preview = req.get("system", "")[:80]
        with open(self.log_file, "a") as f:
            f.write(f"[{ts}] START agent={model} system={system_preview}\n")
        return req

    def on_agent_turn_end(self, response: dict) -> dict:
        ts = datetime.datetime.now().isoformat()
        content_len = len(response.get("content", ""))
        tokens_in = response.get("input_tokens", 0)
        tokens_out = response.get("output_tokens", 0)
        with open(self.log_file, "a") as f:
            f.write(f"[{ts}] END   chars={content_len} in={tokens_in} out={tokens_out}\n")
        return response
