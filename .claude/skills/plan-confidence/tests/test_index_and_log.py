"""T6.1, T6.2, T6.3 — index, log, ADR-001 tests."""
from __future__ import annotations

from pathlib import Path

import yaml


def test_index_lists_plan_confidence_concept(project_root: Path) -> None:
    index = (project_root / ".claude" / "knowledge-base" / "index.md").read_text(encoding="utf-8")
    assert "plan-confidence" in index.lower()
    # Should be linked from SOTA Research section
    pos = index.find("## SOTA Research")
    assert pos != -1
    after = index[pos:pos + 3000]
    assert "Plan-Confidence" in after or "plan-confidence" in after.lower()


def test_log_has_entry_for_v1_1_implementation(project_root: Path) -> None:
    log = (project_root / ".claude" / "knowledge-base" / "log.md").read_text(encoding="utf-8")
    assert "Plan-Confidence v1.1 IMPLEMENTADO" in log or "v1.1 IMPLEMENTADO" in log


def test_log_entry_mentions_6_phases_15_tasks(project_root: Path) -> None:
    log = (project_root / ".claude" / "knowledge-base" / "log.md").read_text(encoding="utf-8")
    # Find the v1.1 entry
    pos = log.find("v1.1 IMPLEMENTADO")
    assert pos != -1
    entry = log[pos:pos + 5000]
    assert "6/6" in entry or "6 phases" in entry.lower()
    assert "15/15" in entry or "15 tasks" in entry.lower()


def test_adr_001_exists(concepts_dir: Path) -> None:
    assert (concepts_dir / "ADR-001-skeleton-and-m2.md").exists()


def test_adr_001_has_frontmatter(concepts_dir: Path) -> None:
    content = (concepts_dir / "ADR-001-skeleton-and-m2.md").read_text(encoding="utf-8")
    assert content.startswith("---\n")
    _, fm_raw, _ = content.split("---\n", 2)
    fm = yaml.safe_load(fm_raw)
    assert fm["adr_id"] == "ADR-plan-conf-001"
    assert fm["status"] in ("proposed", "accepted")


def test_adr_001_has_d1_to_d8(concepts_dir: Path) -> None:
    """v1.1 added D8 (renormalization)."""
    content = (concepts_dir / "ADR-001-skeleton-and-m2.md").read_text(encoding="utf-8")
    for did in ["D1", "D2", "D3", "D4", "D5", "D6", "D7", "D8"]:
        assert f"### {did}" in content, f"missing decision {did}"


def test_adr_001_has_rejected_alternatives(concepts_dir: Path) -> None:
    content = (concepts_dir / "ADR-001-skeleton-and-m2.md").read_text(encoding="utf-8")
    assert "Alternativas Rejeitadas" in content or "Rejected Alternatives" in content
    # At least 3 alternatives documented
    rejected_count = content.count("Rejeitada por")
    assert rejected_count >= 3, f"found {rejected_count} rejected alternatives, expected >= 3"


def test_adr_001_has_consequences_positive_negative_neutral(concepts_dir: Path) -> None:
    content = (concepts_dir / "ADR-001-skeleton-and-m2.md").read_text(encoding="utf-8")
    pos = content.find("## Consequences")
    assert pos != -1
    after = content[pos:]
    assert "Positivas" in after or "Positive" in after
    assert "Negativas" in after or "Negative" in after
    assert "Neutras" in after or "Neutral" in after


def test_adr_001_links_to_plan_and_sota_report(concepts_dir: Path) -> None:
    content = (concepts_dir / "ADR-001-skeleton-and-m2.md").read_text(encoding="utf-8")
    assert "plan-confidence-setup-plan.md" in content
    assert "plan-confidence-sota-" in content
    assert "plan-confidence-golden-rule.md" in content
