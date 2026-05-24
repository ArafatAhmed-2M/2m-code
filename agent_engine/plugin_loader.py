"""
2M Code — Plugin Loader

Discovers, imports, and instantiates plugins from well-known directories.
"""

import importlib
import importlib.util
import logging
import os
import sys

from plugin_base import Plugin

logger = logging.getLogger("2mcode.plugin_loader")

# Directories scanned for plugin files, in priority order.
# On Windows, os.path.expanduser() may return mixed slashes — use normpath.
# The project-local dir is checked at multiple possible CWD locations since
# the engine can be launched from different working directories.
_HOME = os.path.normpath(os.path.expanduser("~"))
_CWD = os.path.normpath(os.getcwd())

PLUGIN_DIRS = [
    os.path.join(_HOME, ".2mcode", "plugins"),
    os.path.join(_CWD, ".2mcode", "plugins"),
    os.path.join(os.path.dirname(_CWD), ".2mcode", "plugins"),
    os.path.join(os.path.dirname(os.path.dirname(_CWD)), ".2mcode", "plugins"),
]


def discover_plugins() -> list[Plugin]:
    """Scan plugin directories and return instantiated Plugin objects.

    Searches PLUGIN_DIRS in order. Each .py file (except __init__.py and
    files starting with _) is imported, and any subclass of Plugin found
    in its module is instantiated.

    Returns:
        A list of instantiated Plugin objects. Plugins earlier in the
        directory list take priority.
    """
    plugins: list[Plugin] = []

    for plugin_dir in PLUGIN_DIRS:
        if not os.path.isdir(plugin_dir):
            logger.debug("Plugin directory not found: %s", plugin_dir)
            continue

        # Collect .py files sorted for deterministic load order
        filenames = sorted(
            f for f in os.listdir(plugin_dir)
            if f.endswith(".py") and not f.startswith("_")
        )

        for filename in filenames:
            filepath = os.path.join(plugin_dir, filename)
            module_name = f"2mcode_plugin_{os.path.splitext(filename)[0]}"

            # Skip already-loaded modules
            if module_name in sys.modules:
                plugin_instances = _extract_plugins(sys.modules[module_name])
                plugins.extend(plugin_instances)
                continue

            try:
                spec = importlib.util.spec_from_file_location(module_name, filepath)
                if spec is None or spec.loader is None:
                    logger.warning("Could not load plugin spec: %s", filepath)
                    continue

                module = importlib.util.module_from_spec(spec)
                sys.modules[module_name] = module
                spec.loader.exec_module(module)

                plugin_instances = _extract_plugins(module)
                plugins.extend(plugin_instances)

            except Exception as e:
                logger.error("Failed to load plugin %s: %s", filepath, e)

    return plugins


def _extract_plugins(module) -> list[Plugin]:
    """Find all Plugin subclasses in a module and instantiate them."""
    instances = []
    for attr_name in dir(module):
        attr = getattr(module, attr_name)
        if (
            isinstance(attr, type)
            and issubclass(attr, Plugin)
            and attr is not Plugin
        ):
            try:
                instance = attr()
                logger.info("Loaded plugin: %s (from %s)", instance.name, getattr(module, "__file__", "?"))
                instances.append(instance)
            except Exception as e:
                logger.error("Failed to instantiate plugin %s: %s", attr_name, e)
    return instances
