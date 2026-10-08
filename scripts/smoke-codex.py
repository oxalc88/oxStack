#!/usr/bin/env python3
"""Read generated MCP configuration with the real Codex CLI in temporary paths."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def run(args, env, cwd=ROOT):
    result = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(f"{args}: {result.stderr}")
    return result.stdout


with tempfile.TemporaryDirectory(prefix="oxstack-codex-") as directory:
    temp = Path(directory).resolve()
    env = dict(os.environ, CODEX_HOME=str(temp / "codex"),
               XDG_STATE_HOME=str(temp / "state"), OXSTACK_ROOT=str(ROOT))
    binary = temp / "oxstack"
    run(["go", "build", "-o", str(binary), "./cmd/oxstack"], env)
    run([str(binary), "install", "--global"], env)
    aws, plain = temp / "aws", temp / "plain"
    for project in (aws, plain):
        project.mkdir()
        run(["git", "init"], env, project)
    config = temp / "codex" / "config.toml"
    config.write_text("".join(
        f"[projects.{json.dumps(str(project))}]\ntrust_level = \"trusted\"\n"
        for project in (aws, plain)))
    for project, selection in ((aws, "aws-lambda"), (plain, "non-aws")):
        run([str(binary), "install", "--project", str(project), "--selection",
             str(ROOT / "selections" / "projects" / f"{selection}.toml")], env)
    def servers(project, profile=None):
        args = ["codex", "-C", str(project)]
        if profile:
            args += ["--profile", profile]
        return json.loads(run(args + ["mcp", "list", "--json"], env, project))
    assert servers(plain) == [], "non-AWS project inherited an MCP"
    aws_servers = {item["name"]: item["enabled"] for item in servers(aws)}
    assert aws_servers == {"aws-documentation": True, "powertools": True}, aws_servers
    run([str(binary), "install", "--profile", "aws-tools"], env)
    assert servers(plain) == [], "profile activated without selection"
    profile_servers = {item["name"] for item in servers(plain, "aws-tools")}
    assert profile_servers == {"aws-pricing", "aws-diagrams"}, profile_servers
    inactive = temp / "inactive.toml"
    inactive.write_text("[mcp.aws-pricing]\ndisabled = true\n[mcp.aws-cloudformation]\nactive = false\n")
    run([str(binary), "install", "--profile", "inactive",
         "--selection", str(inactive)], env)
    disabled = servers(plain, "inactive")
    assert len(disabled) == 2 and all(item["enabled"] is False for item in disabled), disabled
    print("PASS: project isolation, native profiles, inactive MCP, zero base leakage")
    print("No live configuration changed; no MCP servers started.")
