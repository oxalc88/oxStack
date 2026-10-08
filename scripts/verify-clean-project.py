#!/usr/bin/env python3
"""After migration, verify effective MCP definitions and skill prompt catalog."""
import json
from pathlib import Path
import re
import subprocess
import sys

if len(sys.argv) != 2:
    raise SystemExit("Usage: python3 scripts/verify-clean-project.py /non-aws/repo")
project = str(Path(sys.argv[1]).resolve())


def codex(*args):
    result = subprocess.run(["codex", "-C", project, *args],
                            cwd=project, capture_output=True, text=True)
    if result.returncode:
        raise SystemExit(result.stderr)
    return result.stdout


servers = json.loads(codex("mcp", "list", "--json"))
pattern = re.compile(r"aws|bedrock|powertools|serverless", re.I)
leaks = [item["name"] for item in servers if pattern.search(item["name"])]
if leaks:
    raise SystemExit(f"FAIL: AWS MCP definitions remain: {leaks}")
# Inspect the actual model-visible prompt rather than only project skill files.
# The CLI may create ordinary runtime/cache files during prompt inspection.
prompt = codex("debug", "prompt-input")
skills = re.findall(r"[^\s\"<>]*?(?:bedrock|cloudwatch|powertools|aws)[^\s\"<>]*?/SKILL\.md", prompt, re.I)
if skills:
    raise SystemExit(f"FAIL: AWS skill references remain: {skills}")
print("PASS: no AWS MCP definitions or known AWS skill paths in the effective prompt")
print("Also inspect /skills in a fresh desktop thread; account plugins are separate.")
