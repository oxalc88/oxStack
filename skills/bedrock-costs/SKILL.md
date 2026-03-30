---
name: bedrock-costs
context: fork
description: |
  Fetch and analyze AWS Bedrock costs (Cost Explorer + CloudTrail usage attribution).
  Produces a markdown report with per-model, per-user, per-tool, and per-project breakdown.
  Use when asked to "analyze Bedrock costs", "show AWS spending", "cost report",
  "how much are we spending", "Bedrock usage analysis", "analizar costos".
allowed-tools:
  - Bash
  - Read
  - Write
  - AskUserQuestion
---

# AWS Bedrock Cost Analysis

You analyze AWS Bedrock and overall AWS costs using CloudTrail + Cost Explorer.
You produce a structured markdown report with per-model, per-user, per-tool,
and per-project breakdown.

## Step 1: Gather parameters (one interaction)

Ask the user for everything you need in a single `AskUserQuestion`:

```
Para analizar los costos de AWS, necesito confirmar los parámetros:

1. **AWS Profile**: ¿Qué perfil usar?
   (detectado: $AWS_PROFILE={value_or_empty})
2. **Período**: ¿Desde qué fecha? (default: primer día del mes actual)
3. **Output dir**: ¿Dónde guardar los JSONs? (default: ./bedrock-costs)
4. **Tag de proyecto**: ¿Qué tag usar para desglosar por proyecto? (default: project)

¿Confirmas estos valores o quieres cambiar algo?
```

Once confirmed, proceed without further questions.

## Step 2: Verify AWS access

```bash
aws sts get-caller-identity --profile <profile>
```

If this fails, explain the error and stop.

## Step 3: Fetch data

The Go binaries live in the skill's own `scripts/bin/` directory.
Auto-build if needed, then run fetch:

```bash
SKILL_DIR=$(readlink -f ~/.claude/skills/bedrock-costs)
GO_DIR="$SKILL_DIR/scripts"
BIN_DIR="$GO_DIR/bin"

if [ ! -f "$BIN_DIR/fetch" ] || [ "$(find "$GO_DIR" -name '*.go' -newer "$BIN_DIR/fetch" 2>/dev/null)" ]; then
  mkdir -p "$BIN_DIR"
  (cd "$GO_DIR" && mise exec -- go build -o "$BIN_DIR/fetch" ./cmd/fetch && mise exec -- go build -o "$BIN_DIR/analyze" ./cmd/analyze)
fi

"$BIN_DIR/fetch" \
  --profile <profile> \
  --start <start_date> \
  --end <end_date> \
  --output-dir <output_dir> \
  --regions us-east-1,us-west-2,eu-west-1,ap-southeast-1,us-east-2 \
  --tag-key <tag_key>
```

## Step 4: Analyze and generate report

```bash
"$BIN_DIR/analyze" \
  --data-dir <output_dir> \
  --format both \
  --output <output_dir>/report.md
```

## Step 5: Present findings with interpretation

Read the generated `report.md` and present a concise summary with interpretation:

- **Total cost and period**
- **Top cost drivers** (models, users, or tools dominating spend)
- **Notable spikes** (days where cost was >2× the median — already flagged in the report)
- **Caching impact**: if `us.*` models are used heavily, recommend switching to `global.*`
- **Tools breakdown**: which dev tools (Claude Code, KiloCode, OpenCode, Cline) dominate
- **Deployed apps**: identify Lambda/service identities vs human users
- **Recommendations**: top 2-3 actionable items to reduce cost

Keep the interpretation to ~10 bullet points. The full detail is in the report.

## Error handling

- **No CloudTrail events**: region may have no Bedrock activity, or the date range is outside the 90-day CloudTrail window. Mention this limitation.
- **Tag query fails with AccessDeniedException**: normal for linked accounts — the tag cost breakdown will be empty. Explain this in the report.
- **Build fails**: ensure `mise` and Go 1.26 are available (`mise exec -- go version`). The `scripts/mise.toml` pins Go 1.26.
- **Credential errors**: suggest `aws sso login --profile <profile>` or checking `~/.aws/credentials`.

## Key behaviors

- **One question, then autonomous execution** — do not ask for confirmation mid-run.
- **Always generate the markdown report** even if CloudTrail returns 0 events.
- **Interpret, don't just report** — Claude adds context (e.g., "the Feb 22 spike correlates with KiloCode switching to Opus 4.6 without caching").
- **Portable**: works with any AWS profile and any account, not just OxBedrock.
