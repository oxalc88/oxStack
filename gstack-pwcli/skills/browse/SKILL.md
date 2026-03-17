---
name: browse
description: Browser automation for QA, dogfooding, and web inspection using playwright-cli. Use when the user needs to navigate websites, test web applications, fill forms, take screenshots, inspect console errors, or verify deployments. Replaces gstack's compiled browse binary with Microsoft's playwright-cli — no Bun required, works on Ubuntu and macOS.
allowed-tools: Bash(playwright-cli:*), Bash(npx playwright-cli:*)
---

# Browser Automation with playwright-cli

## Prerequisites

```bash
npm install -g @playwright/mcp@latest
```

If the global binary is not available, fall back to `npx playwright-cli`.

## Quick start

```bash
# open browser and navigate
playwright-cli open https://staging.myapp.com

# get accessibility snapshot (element refs like e1, e2, e3...)
playwright-cli snapshot

# interact using refs from the snapshot
playwright-cli fill e2 "test@example.com"
playwright-cli fill e3 "password123"
playwright-cli click e5

# screenshot for visual verification
playwright-cli screenshot

# check for JS errors
playwright-cli console
```

## Core workflow

The browsing workflow is: navigate → snapshot → interact → re-snapshot.

1. **Navigate**: `playwright-cli open <url>`
2. **Snapshot**: `playwright-cli snapshot` to get the accessibility tree with element refs
3. **Interact**: use refs from the snapshot to click, fill, type, select
4. **Re-snapshot**: after significant changes (page navigation, form submission, modal open) take a new snapshot to get fresh refs
5. **Verify**: use `playwright-cli screenshot` for visual checks, `playwright-cli console` for JS errors

Always snapshot before interacting. Refs become stale after page mutations.

## Command reference

### Core

```bash
playwright-cli open <url>                # open browser + navigate (first call starts Chromium)
playwright-cli close                     # close the page
playwright-cli type "search query"       # type text into focused/editable element
playwright-cli click <ref>               # click element by ref (e.g. e5)
playwright-cli dblclick <ref>            # double-click element
playwright-cli fill <ref> "text"         # fill input field by ref
playwright-cli drag <startRef> <endRef>  # drag and drop between two elements
playwright-cli hover <ref>               # hover over element
playwright-cli select <ref> "value"      # select dropdown option
playwright-cli upload ./file.pdf         # upload file to file input
playwright-cli check <ref>               # check checkbox or radio
playwright-cli uncheck <ref>             # uncheck checkbox
playwright-cli snapshot                  # accessibility tree with element refs
playwright-cli eval "document.title"     # evaluate JS expression on page
playwright-cli eval "el => el.textContent" <ref>  # evaluate JS on specific element
playwright-cli dialog-accept             # accept browser dialog
playwright-cli dialog-dismiss            # dismiss browser dialog
playwright-cli resize 1920 1080          # resize viewport
```

### Navigation

```bash
playwright-cli go-back                   # browser back
playwright-cli go-forward                # browser forward
playwright-cli reload                    # reload current page
```

### Keyboard

```bash
playwright-cli press Enter               # press key
playwright-cli press ArrowDown
playwright-cli keydown Shift
playwright-cli keyup Shift
```

### Mouse (low-level)

```bash
playwright-cli mousemove 150 300
playwright-cli mousedown
playwright-cli mouseup
playwright-cli mousewheel 0 100          # scroll down
```

### Screenshots & Export

```bash
playwright-cli screenshot                # full page screenshot
playwright-cli screenshot <ref>          # screenshot specific element
playwright-cli pdf                       # save page as PDF
```

### Tabs

```bash
playwright-cli tab-list                  # list all open tabs
playwright-cli tab-new                   # open new empty tab
playwright-cli tab-new https://url       # open new tab with URL
playwright-cli tab-close                 # close current tab
playwright-cli tab-close 2               # close tab by index
playwright-cli tab-select 0              # switch to tab by index
```

### DevTools

```bash
playwright-cli console                   # all console messages
playwright-cli console warning           # filter by level (warning, error)
playwright-cli network                   # list network requests since page load
playwright-cli run-code "await page.waitForTimeout(1000)"  # run arbitrary Playwright code
playwright-cli tracing-start             # start trace recording
playwright-cli tracing-stop              # stop trace and save trace file
```

### Sessions

Sessions isolate browser state (cookies, localStorage, profiles) per project.
Use sessions to test multiple environments without cross-contamination.

```bash
# start a named session
playwright-cli --session=staging open https://staging.myapp.com

# interact within that session
playwright-cli --session=staging click e6
playwright-cli --session=staging snapshot

# manage sessions
playwright-cli session-list              # list active sessions
playwright-cli session-stop staging      # stop specific session
playwright-cli session-stop-all          # stop all sessions
playwright-cli session-delete staging    # delete session data + profile
```

You can also set the session via environment variable to avoid repeating `--session`:

```bash
export PLAYWRIGHT_CLI_SESSION=staging
playwright-cli open https://staging.myapp.com
playwright-cli snapshot
```

## Command mapping from gstack browse

If you are used to gstack's `/browse` commands, here is the translation:

| gstack browse              | playwright-cli equivalent              |
|----------------------------|----------------------------------------|
| `browse goto <url>`        | `playwright-cli open <url>`            |
| `browse snapshot`          | `playwright-cli snapshot`              |
| `browse snapshot -i`       | `playwright-cli snapshot`              |
| `browse fill @e2 "text"`  | `playwright-cli fill e2 "text"`        |
| `browse click @e5`         | `playwright-cli click e5`              |
| `browse screenshot <path>` | `playwright-cli screenshot`            |
| `browse console`           | `playwright-cli console`               |
| `browse text`              | `playwright-cli eval "document.body.innerText"` |
| `browse tabs`              | `playwright-cli tab-list`              |

Key differences:
- No `@` prefix on element refs — use `e5` not `@e5`
- Screenshots save automatically (no path argument needed)
- No compiled binary — `playwright-cli` is a standard npm package
- Sessions are managed with `--session=name` flag or `PLAYWRIGHT_CLI_SESSION` env var
- No Bun dependency — works with Node.js only

## QA workflow patterns

### Quick smoke test

Verify a deployment is alive — homepage loads, no console errors, key pages reachable:

```bash
playwright-cli open https://staging.myapp.com
playwright-cli console
playwright-cli screenshot
playwright-cli open https://staging.myapp.com/dashboard
playwright-cli console
playwright-cli screenshot
```

### Form flow testing

Test a signup or submission flow end to end:

```bash
playwright-cli open https://myapp.com/signup
playwright-cli snapshot
# identify form fields from snapshot refs
playwright-cli fill e1 "test@example.com"
playwright-cli fill e2 "StrongPass123!"
playwright-cli click e5                    # submit button
playwright-cli snapshot                    # verify redirect / success state
playwright-cli console                     # check for errors
playwright-cli screenshot                  # visual proof
```

### Multi-page verification

After a branch push, check all changed routes:

```bash
# use a named session so cookies persist across pages
playwright-cli --session=qa open https://staging.myapp.com/login
playwright-cli --session=qa snapshot
playwright-cli --session=qa fill e2 "admin@test.com"
playwright-cli --session=qa fill e3 "password"
playwright-cli --session=qa click e5

# now test each affected route
playwright-cli --session=qa open https://staging.myapp.com/listings/new
playwright-cli --session=qa snapshot
playwright-cli --session=qa console
playwright-cli --session=qa screenshot

playwright-cli --session=qa open https://staging.myapp.com/listings/123
playwright-cli --session=qa snapshot
playwright-cli --session=qa console
playwright-cli --session=qa screenshot
```

### Debugging with traces

When something is broken and you need detailed diagnostics:

```bash
playwright-cli open https://myapp.com/broken-page
playwright-cli tracing-start
playwright-cli click e4
playwright-cli fill e7 "test"
playwright-cli tracing-stop
# trace file is saved — can be opened with Playwright Trace Viewer
```

## Authentication

playwright-cli sessions are persistent — cookies and localStorage carry over between
calls within the same session. To test authenticated pages:

1. Start a named session: `playwright-cli --session=myapp open https://myapp.com/login`
2. Log in through the UI using `fill` + `click` commands
3. All subsequent calls with `--session=myapp` will have the session cookies

This replaces gstack's `/setup-browser-cookies` macOS-only cookie importer with a
simpler, cross-platform approach that works on both Ubuntu and macOS.

## Headless vs Headed

playwright-cli runs headless by default (ideal for Ubuntu servers, CI, and remote machines).
To see the browser on macOS for debugging:

```bash
playwright-cli open https://myapp.com --headed
```

## Troubleshooting

**`playwright-cli: command not found`**
→ Install globally: `npm install -g @playwright/mcp@latest`
→ Or use `npx playwright-cli` as fallback

**Browser not launching on Ubuntu**
→ Install system dependencies: `npx playwright install-deps chromium`
→ Then install browser: `npx playwright install chromium`

**Stale element refs**
→ Always re-snapshot after page navigation, form submissions, or modal interactions.
   Refs from a previous snapshot point to the old DOM state.

**Session cleanup**
→ `playwright-cli session-stop-all` stops all running browsers
→ `playwright-cli session-delete <name>` removes stored profile data
