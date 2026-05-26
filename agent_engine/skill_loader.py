"""
2M Code — Skill Loader

Scans Skills/ directories for SKILL.md files, parses YAML frontmatter + body.
Skills are collections of structured instructions that can be injected into
agent system prompts based on user request context.
"""

import logging
import os
from pathlib import Path
from typing import Optional
import yaml

logger = logging.getLogger("2mcode.skills")

SKILL_DIR_NAME = "Skills"
SKILL_FILE_NAME = "SKILL.md"


def _find_skills_root() -> Optional[Path]:
    """Locate the Skills/ directory relative to the agent_engine package."""
    # agent_engine/ is a sibling of Skills/
    engine_dir = Path(__file__).parent.resolve()
    candidate = engine_dir.parent / SKILL_DIR_NAME
    if candidate.is_dir():
        return candidate
    # Also check installed layout: ~/.2mcode/Skills/
    home = Path.home()
    candidate = home / ".2mcode" / SKILL_DIR_NAME
    if candidate.is_dir():
        return candidate
    return None


def _parse_skill_file(path: Path) -> Optional[dict]:
    """Parse a single SKILL.md file into name, description, license, content."""
    try:
        text = path.read_text(encoding="utf-8")
    except Exception as e:
        logger.warning("Cannot read skill file %s: %s", path, e)
        return None

    # Extract YAML frontmatter between --- delimiters
    content = text.strip()
    if not content.startswith("---"):
        logger.warning("Skill file %s missing YAML frontmatter", path)
        return None

    # Find second ---
    end_idx = content.find("---", 3)
    if end_idx == -1:
        logger.warning("Skill file %s has unclosed frontmatter", path)
        return None

    frontmatter_text = content[3:end_idx].strip()
    body = content[end_idx + 3:].strip()

    try:
        meta = yaml.safe_load(frontmatter_text)
    except yaml.YAMLError as e:
        logger.warning("Skill file %s has invalid YAML: %s", path, e)
        return None

    if not isinstance(meta, dict):
        return None

    name = (meta.get("name") or "").strip()
    if not name:
        logger.warning("Skill file %s missing 'name' in frontmatter", path)
        return None

    return {
        "name": name,
        "description": (meta.get("description") or "").strip(),
        "license": (meta.get("license") or "").strip(),
        "content": body,
        "path": str(path),
    }


def list_skills() -> list[dict]:
    """Scan the Skills/ directory and return metadata for all skills."""
    root = _find_skills_root()
    if root is None:
        logger.warning("Skills/ directory not found")
        return []

    skills = []
    for entry in sorted(root.iterdir()):
        if not entry.is_dir():
            continue
        skill_file = entry / SKILL_FILE_NAME
        if not skill_file.is_file():
            continue
        parsed = _parse_skill_file(skill_file)
        if parsed is not None:
            skills.append(parsed)

    return skills


def get_skill(name: str) -> Optional[dict]:
    """Get a single skill by name."""
    for skill in list_skills():
        if skill["name"] == name:
            return skill
    return None
