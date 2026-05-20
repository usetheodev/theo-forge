"""T3.1 + T3.2 — concepts/plan-confidence/ skeleton + holdout-entry-template."""
from __future__ import annotations

from pathlib import Path

import yaml


def _read_md(path: Path) -> tuple[dict[str, object], str]:
    content = path.read_text(encoding="utf-8")
    if not content.startswith("---\n"):
        return {}, content
    _, fm_raw, body = content.split("---\n", 2)
    fm: dict[str, object] = yaml.safe_load(fm_raw) or {}
    return fm, body


# T3.1 — Concepts skeleton

def test_concepts_plan_confidence_dir_exists(concepts_dir: Path) -> None:
    assert concepts_dir.exists() and concepts_dir.is_dir()


def test_concepts_index_readme_has_frontmatter(concepts_dir: Path) -> None:
    fm, _ = _read_md(concepts_dir / "README.md")
    assert "type" in fm
    assert "created_at" in fm


def test_holdout_readme_exists(concepts_dir: Path) -> None:
    assert (concepts_dir / "holdout" / "README.md").exists()


def test_outcomes_readme_exists(concepts_dir: Path) -> None:
    assert (concepts_dir / "outcomes" / "README.md").exists()


def test_index_readme_links_to_holdout_and_outcomes(concepts_dir: Path) -> None:
    _, body = _read_md(concepts_dir / "README.md")
    assert "holdout" in body.lower()
    assert "outcomes" in body.lower()


def test_holdout_readme_explains_format(concepts_dir: Path) -> None:
    content = (concepts_dir / "holdout" / "README.md").read_text(encoding="utf-8")
    content_lower = content.lower()
    assert "template" in content_lower or "formato" in content_lower or "format" in content_lower
    assert "graded" in content_lower or "score" in content_lower


def test_outcomes_readme_explains_format(concepts_dir: Path) -> None:
    content = (concepts_dir / "outcomes" / "README.md").read_text(encoding="utf-8")
    content_lower = content.lower()
    assert "pass" in content_lower or "fail" in content_lower or "outcome" in content_lower


def test_concepts_readmes_under_200_lines(concepts_dir: Path) -> None:
    for path in [
        concepts_dir / "README.md",
        concepts_dir / "holdout" / "README.md",
        concepts_dir / "outcomes" / "README.md",
    ]:
        lines = len(path.read_text(encoding="utf-8").splitlines())
        assert lines <= 200, f"{path.name} has {lines} lines (max 200)"


# T3.2 — Holdout entry template

def test_holdout_entry_template_exists(skill_root: Path) -> None:
    assert (skill_root / "templates" / "holdout-entry-template.md").exists()


def test_holdout_entry_template_has_required_frontmatter_fields(skill_root: Path) -> None:
    content = (skill_root / "templates" / "holdout-entry-template.md").read_text(encoding="utf-8")
    # The template SHOWS the frontmatter fields users must fill.
    required = ["plan_path", "plan_slug", "graded_by", "graded_at", "type"]
    for field in required:
        assert field in content, f"template missing field hint: {field}"


def test_holdout_entry_template_has_4_dimension_sections(skill_root: Path) -> None:
    content = (skill_root / "templates" / "holdout-entry-template.md").read_text(encoding="utf-8")
    content_lower = content.lower()
    assert "completude" in content_lower
    assert "evidencia" in content_lower or "evidência" in content_lower
    assert "calibraca" in content_lower or "calibração" in content_lower or "calibracao" in content_lower
    assert "risco" in content_lower


def test_holdout_entry_template_has_hard_caps_checklist(skill_root: Path) -> None:
    content = (skill_root / "templates" / "holdout-entry-template.md").read_text(encoding="utf-8")
    assert "Hard Caps" in content or "hard caps" in content.lower()
    # Should reference Coverage Matrix, ADR alternatives, TDD
    content_lower = content.lower()
    assert "coverage" in content_lower
    assert "tdd" in content_lower or "regression test" in content_lower


def test_holdout_entry_template_has_outcome_section(skill_root: Path) -> None:
    content = (skill_root / "templates" / "holdout-entry-template.md").read_text(encoding="utf-8")
    content_lower = content.lower()
    assert "outcome" in content_lower
    assert "cross_validation" in content or "cross-validation" in content_lower
    assert "dogfood" in content_lower
