---
name: browse
description: Browser automation for QA, dogfooding, and web inspection using playwright-mcp. Use when the user needs to navigate websites, test web applications, fill forms, take screenshots, inspect console errors, or verify deployments. Replaces gstack's compiled browse binary with Microsoft's playwright-mcp — no Bun required, works on Ubuntu and macOS.
allowed-tools: Bash(playwright-mcp:*), Bash(npx playwright-mcp:*)
---

# Browser Automation with playwright-mcp

## Prerequisites

```bash
npm install -g @playwright/mcp@latest
```

If the global binary is not available, fall back to `npx playwright-mcp`.

## Quick start

```bash
# open browser and navigate
playwright-mcp open https://staging.myapp.com

# get accessibility snapshot (element refs like e1, e2, e3...)
playwright-mcp snapshot

# interact using refs from the snapshot
playwright-mcp fill e2 "test@example.com"
playwright-mcp fill e3 "password123"
playwright-mcp click e5

# screenshot for visual verification
playwright-mcp screenshot

# check for JS errors
playwright-mcp console
```

## Core workflow

The browsing workflow is: navigate → snapshot → interact → re-snapshot.

1. **Navigate**: `playwright-mcp open <url>`
2. **Snapshot**: `playwright-mcp snapshot` to get the accessibility tree with element refs
3. **Interact**: use refs from the snapshot to click, fill, type, select
4. **Re-snapshot**: after significant changes (page navigation, form submission, modal open) take a new snapshot to get fresh refs
5. **Verify**: use `playwright-mcp screenshot` for visual checks, `playwright-mcp console` for JS errors

Always snapshot before interacting. Refs become stale after page mutations.

## Command reference

### Core

```bash
playwright-mcp open <url>                # open browser + navigate (first call starts Chromium)
playwright-mcp close                     # close the page
playwright-mcp type "search query"       # type text into focused/editable element
playwright-mcp click <ref>               # click element by ref (e.g. e5)
playwright-mcp dblclick <ref>            # double-click element
playwright-mcp fill <ref> "text"         # fill input field by ref
playwright-mcp drag <startRef> <endRef>  # drag and drop between two elements
playwright-mcp hover <ref>               # hover over element
playwright-mcp select <ref> "value"      # select dropdown option
playwright-mcp upload ./file.pdf         # upload file to file input
playwright-mcp check <ref>               # check checkbox or radio
playwright-mcp uncheck <ref>             # uncheck checkbox
playwright-mcp snapshot                  # accessibility tree with element refs
playwright-mcp eval "document.title"     # evaluate JS expression on page
playwright-mcp eval "el => el.textContent" <ref>  # evaluate JS on specific element
playwright-mcp dialog-accept             # accept browser dialog
playwright-mcp dialog-dismiss            # dismiss browser dialog
playwright-mcp resize 1920 1080          # resize viewport
```

### Navigation

```bash
playwright-mcp go-back                   # browser back
playwright-mcp go-forward                # browser forward
playwright-mcp reload                    # reload current page
```

### Keyboard

```bash
playwright-mcp press Enter               # press key
playwright-mcp press ArrowDown
playwright-mcp keydown Shift
playwright-mcp keyup Shift
```

### Mouse (low-level)

```bash
playwright-mcp mousemove 150 300
playwright-mcp mousedown
playwright-mcp mouseup
playwright-mcp mousewheel 0 100          # scroll down
```

### Screenshots & Export

```bash
playwright-mcp screenshot                # full page screenshot
playwright-mcp screenshot <ref>          # screenshot specific element
playwright-mcp pdf                       # save page as PDF
```

### Tabs

```bash
playwright-mcp tab-list                  # list all open tabs
playwright-mcp tab-new                   # open new empty tab
playwright-mcp tab-new https://url       # open new tab with URL
playwright-mcp tab-close                 # close current tab
playwright-mcp tab-close 2               # close tab by index
playwright-mcp tab-select 0              # switch to tab by index
```

### DevTools

```bash
playwright-mcp console                   # all console messages
playwright-mcp console warning           # filter by level (warning, error)
playwright-mcp network                   # list network requests since page load
playwright-mcp run-code "await page.waitForTimeout(1000)"  # run arbitrary Playwright code
playwright-mcp tracing-start             # start trace recording
playwright-mcp tracing-stop              # stop trace and save trace file
```

### Sessions

Sessions isolate browser state (cookies, localStorage, profiles) per project.
Use sessions to test multiple environments without cross-contamination.

```bash
# start a named session
playwright-mcp --session=staging open https://staging.myapp.com

# interact within that session
playwright-mcp --session=staging click e6
playwright-mcp --session=staging snapshot

# manage sessions
playwright-mcp session-list              # list active sessions
playwright-mcp session-stop staging      # stop specific session
playwright-mcp session-stop-all          # stop all sessions
playwright-mcp session-delete staging    # delete session data + profile
```

You can also set the session via environment variable to avoid repeating `--session`:

```bash
export PLAYWRIGHT_CLI_SESSION=staging
playwright-mcp open https://staging.myapp.com
playwright-mcp snapshot
```

## Command mapping from gstack browse

If you are used to gstack's `/browse` commands, here is the translation:

| gstack browse              | playwright-mcp equivalent              |
|----------------------------|----------------------------------------|
| `browse goto <url>`        | `playwright-mcp open <url>`            |
| `browse snapshot`          | `playwright-mcp snapshot`              |
| `browse snapshot -i`       | `playwright-mcp snapshot`              |
| `browse fill @e2 "text"`  | `playwright-mcp fill e2 "text"`        |
| `browse click @e5`         | `playwright-mcp click e5`              |
| `browse screenshot <path>` | `playwright-mcp screenshot`            |
| `browse console`           | `playwright-mcp console`               |
| `browse text`              | `playwright-mcp eval "document.body.innerText"` |
| `browse tabs`              | `playwright-mcp tab-list`              |

Key differences:
- No `@` prefix on element refs — use `e5` not `@e5`
- Screenshots save automatically (no path argument needed)
- No compiled binary — `playwright-mcp` is a standard npm package
- Sessions are managed with `--session=name` flag or `PLAYWRIGHT_CLI_SESSION` env var
- No Bun dependency — works with Node.js only

## QA workflow patterns

### Quick smoke test

Verify a deployment is alive — homepage loads, no console errors, key pages reachable:

```bash
playwright-mcp open https://staging.myapp.com
playwright-mcp console
playwright-mcp screenshot
playwright-mcp open https://staging.myapp.com/dashboard
playwright-mcp console
playwright-mcp screenshot
```

### Form flow testing

Test a signup or submission flow end to end:

```bash
playwright-mcp open https://myapp.com/signup
playwright-mcp snapshot
# identify form fields from snapshot refs
playwright-mcp fill e1 "test@example.com"
playwright-mcp fill e2 "StrongPass123!"
playwright-mcp click e5                    # submit button
playwright-mcp snapshot                    # verify redirect / success state
playwright-mcp console                     # check for errors
playwright-mcp screenshot                  # visual proof
```

### Multi-page verification

After a branch push, check all changed routes:

```bash
# use a named session so cookies persist across pages
playwright-mcp --session=qa open https://staging.myapp.com/login
playwright-mcp --session=qa snapshot
playwright-mcp --session=qa fill e2 "admin@test.com"
playwright-mcp --session=qa fill e3 "password"
playwright-mcp --session=qa click e5

# now test each affected route
playwright-mcp --session=qa open https://staging.myapp.com/listings/new
playwright-mcp --session=qa snapshot
playwright-mcp --session=qa console
playwright-mcp --session=qa screenshot

playwright-mcp --session=qa open https://staging.myapp.com/listings/123
playwright-mcp --session=qa snapshot
playwright-mcp --session=qa console
playwright-mcp --session=qa screenshot
```

### Debugging with traces

When something is broken and you need detailed diagnostics:

```bash
playwright-mcp open https://myapp.com/broken-page
playwright-mcp tracing-start
playwright-mcp click e4
playwright-mcp fill e7 "test"
playwright-mcp tracing-stop
# trace file is saved — can be opened with Playwright Trace Viewer
```

## Authentication

playwright-mcp sessions are persistent — cookies and localStorage carry over between
calls within the same session. To test authenticated pages:

1. Start a named session: `playwright-mcp --session=myapp open https://myapp.com/login`
2. Log in through the UI using `fill` + `click` commands
3. All subsequent calls with `--session=myapp` will have the session cookies

This replaces gstack's `/setup-browser-cookies` macOS-only cookie importer with a
simpler, cross-platform approach that works on both Ubuntu and macOS.

## Headless vs Headed

playwright-mcp runs headless by default (ideal for Ubuntu servers, CI, and remote machines).
To see the browser on macOS for debugging:

```bash
playwright-mcp open https://myapp.com --headed
```

## Troubleshooting

**`playwright-mcp: command not found`**
→ Install globally: `npm install -g @playwright/mcp@latest`
→ Or use `npx playwright-mcp` as fallback

**Browser not launching on Ubuntu**
→ Install system dependencies: `npx playwright install-deps chromium`
→ Then install browser: `npx playwright install chromium`

**Stale element refs**
→ Always re-snapshot after page navigation, form submissions, or modal interactions.
   Refs from a previous snapshot point to the old DOM state.

**Session cleanup**
→ `playwright-mcp session-stop-all` stops all running browsers
→ `playwright-mcp session-delete <name>` removes stored profile data
