"""T5.1 + post-MANDATORY upgrade — /to-plan chain integration tests.

After the chain was promoted from RECOMMENDED to MANDATORY, the SKILL.md
must:
- Use the word MANDATORY (in any case) to describe the chain
- Reference all 6 steps in order
- Include /plan-improve as Step 5
- Document stopping conditions
- Link to plan-confidence + plan-improve + golden-rule
"""
from __future__ import annotations

from pathlib import Path


def _read(project_root: Path) -> str:
    return (project_root / ".claude" / "skills" / "to-plan" / "SKILL.md").read_text(encoding="utf-8")


def test_to_plan_skill_mentions_plan_confidence(project_root: Path) -> None:
    content = _read(project_root)
    assert "plan-confidence" in content.lower()


def test_to_plan_skill_mentions_plan_improve(project_root: Path) -> None:
    content = _read(project_root)
    assert "plan-improve" in content.lower()


def test_chain_is_mandatory_not_recommended(project_root: Path) -> None:
    """Chain must be MANDATORY. RECOMMENDED was the old design."""
    content = _read(project_root)
    text_lower = content.lower()
    # MANDATORY must appear in the post-plan section
    assert "mandatory" in text_lower
    # Old "RECOMMENDED" verbiage for the chain should not be there
    # (it can appear in other contexts, but not as the chain's status).
    # We check for the specific section title.
    assert "(recommended)" not in content.lower() or "post-edge-case: plan confidence scoring (recommended)" not in content.lower()


def test_chain_has_all_six_steps(project_root: Path) -> None:
    """All 6 mandatory steps must be present in order."""
    content = _read(project_root)
    # Step labels (loose matching)
    expected_phrases_in_order = [
        "step 1",  # Generate plan
        "step 2",  # /edge-case-plan
        "step 3",  # Incorporate MUST FIX
        "step 4",  # /plan-confidence
        "step 5",  # /plan-improve
        "step 6",  # /plan-confidence re-score
    ]
    last_pos = -1
    for phrase in expected_phrases_in_order:
        pos = content.lower().find(phrase)
        assert pos != -1, f"missing chain step: {phrase}"
        assert pos > last_pos, f"step out of order: {phrase} at {pos}, last at {last_pos}"
        last_pos = pos


def test_chain_documents_pipeline_contract(project_root: Path) -> None:
    content = _read(project_root)
    text_lower = content.lower()
    assert "pipeline contract" in text_lower or "chain inquebr" in text_lower


def test_chain_documents_stopping_conditions(project_root: Path) -> None:
    content = _read(project_root)
    text_lower = content.lower()
    assert "stopping condition" in text_lower or "exhausted" in text_lower or "max-iterations" in text_lower


def test_chain_documents_hard_caps(project_root: Path) -> None:
    """Hard caps must be referenced in the chain doc."""
    content = _read(project_root)
    assert "Coverage Matrix" in content
    assert "49" in content
    assert "70" in content


def test_chain_links_to_plan_confidence_skill(project_root: Path) -> None:
    content = _read(project_root)
    assert "plan-confidence/SKILL.md" in content


def test_chain_links_to_plan_improve_skill(project_root: Path) -> None:
    content = _read(project_root)
    assert "plan-improve/SKILL.md" in content


def test_chain_links_to_golden_rule(project_root: Path) -> None:
    content = _read(project_root)
    assert "plan-confidence-golden-rule.md" in content


def test_chain_documents_re_score_after_improve(project_root: Path) -> None:
    """Step 6: re-score after /plan-improve is the audit trail."""
    content = _read(project_root)
    text_lower = content.lower()
    assert "re-score" in text_lower or "re-run" in text_lower or "audit trail" in text_lower


def test_chain_explicitly_supersedes_old_d6(project_root: Path) -> None:
    """ADR D6 (chain optional) is now SUPERSEDED by MANDATORY."""
    content = _read(project_root)
    text_lower = content.lower()
    assert "superseded" in text_lower or "d6" in text_lower


def test_chain_does_not_allow_skip(project_root: Path) -> None:
    """The chain MUST NOT have language like 'can be skipped' or 'optional during rollout'."""
    content = _read(project_root)
    text_lower = content.lower()
    # These OLD phrases must be gone or scoped to historical context
    assert "can be skipped during rollout" not in text_lower
    # "RECOMMENDED but not mandatory" was the old verbiage; must not appear
    assert "recommended but not mandatory" not in text_lower


def test_chain_emits_user_facing_summary(project_root: Path) -> None:
    """The chain ends with a user-facing summary (audit trail + next instruction)."""
    content = _read(project_root)
    text_lower = content.lower()
    assert "audit trail" in text_lower
    assert "what the user sees" in text_lower or "user invokes" in text_lower
