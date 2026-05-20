"""T5.2 — CLAUDE.md mentions /plan-confidence in skills table."""
from __future__ import annotations

from pathlib import Path


def test_claude_md_mentions_plan_confidence(project_root: Path) -> None:
    content = (project_root / "CLAUDE.md").read_text(encoding="utf-8")
    # Should appear in the "Skills that write" table
    pos = content.find("Skills that write")
    assert pos != -1
    after = content[pos:pos + 3000]  # enough window
    assert "/plan-confidence" in after


def test_claude_md_skill_table_still_valid_markdown(project_root: Path) -> None:
    """Sanity check: the table around plan-confidence still has pipes."""
    content = (project_root / "CLAUDE.md").read_text(encoding="utf-8")
    line_index = None
    for i, line in enumerate(content.splitlines()):
        if "plan-confidence" in line and line.startswith("|"):
            line_index = i
            break
    assert line_index is not None
    # That line has 4 cells (3 pipes + leading/trailing)
    lines = content.splitlines()
    cells = lines[line_index].split("|")
    assert len(cells) >= 4, f"row has {len(cells)} cells: {lines[line_index]!r}"
