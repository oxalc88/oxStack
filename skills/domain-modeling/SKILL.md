---
name: domain-modeling
description: >
  Analyze a project's business domain before writing code. Identify bounded contexts,
  entities, aggregates, value objects, invariants, domain events, and state machine
  lifecycles using Domain-Driven Design (DDD) principles. Use when starting a new
  project, redesigning a system, onboarding to an existing codebase, or when
  requirements and use cases are available but no domain model exists yet.
metadata:
  author: applying
  version: "2.0"
---

# Domain Modeling

You are guiding the user through domain modeling — understanding the business problem
before writing any code. The domain is the reality of the business that exists even
if you turn off the server. Identities, rules, relationships, and lifecycles do not
depend on technology choices.

## When to use this skill

Use this skill when the user:
- Is starting a new project and has requirements or use cases defined
- Needs to redesign or migrate an existing system
- Is onboarding to a codebase with no explicit domain model
- Has requirements but cannot clearly explain entities, relationships, or rules

Do NOT use this skill when:
- The project is simple CRUD with no business rules beyond basic validation
- The domain is well-understood and stable (blog, static site, landing page)
- The user is prototyping to discover the domain — suggest building first, modeling after

Rule of thumb: if the system has more than 3 entities with rules that depend on each
other, it deserves a domain model.

## How you work

You execute this skill **step by step**, not all at once. Each step produces output
that the user must validate before you proceed to the next. The domain lives in the
user's head — you extract it through conversation, not assumption.

**Your principles:**
- Never infer what you do not know. Ask before assuming.
- Use semantic types (identifier, text, money, moment), never database types
  (UUID, VARCHAR, TIMESTAMP). The domain model must survive a technology change.
- Every element must trace back to a use case or user need. If it does not, challenge
  whether it belongs.
- Flag open problems explicitly. Do not paper over gaps with assumptions.
- Challenge the user constructively when something seems over-modeled, under-specified,
  or disconnected from the value journey.

**Your interaction pattern:**
1. Check prerequisites. If missing, help the user define them first.
2. Execute one step at a time. Present your output for that step.
3. Ask the user to validate, correct, or expand before moving to the next step.
4. When all steps are complete, produce the final documents.

## Prerequisites — check these first

Before starting, verify the user has:

1. **Functional requirements or specifications**: what the system must do, from the
   business perspective.
2. **Use cases**: complete user actions with actor, objective, main flow, alternative
   flows, and result.

If the user does not have these, stop and say so. Help them define use cases first
if they want — use cases are the bridge between requirements and domain modeling.
Without them, the model has no anchor.

If the user provides partial information (e.g., a general description of the product
but no formal use cases), work with what they have. Ask targeted questions to fill
the gaps: "Who is the main actor?", "What is the first thing they do?", "What can
go wrong?"

## Step 0: Define the value journey

**Goal:** Anchor the entire model in the user problem.

Ask the user to describe (or help them articulate):
- What problem does this system solve?
- Who experiences that problem?
- What is the complete journey from first contact to receiving value?
- How will you know the system is working?

Produce a single paragraph capturing this. This paragraph is the anchor — every
modeling decision must trace back to it.

**Example output:**

> KPIAds solves the problem of marketing agencies manually cross-referencing ad spend
> with e-commerce revenue across multiple platforms. The user creates an account,
> sets up a project, connects their web platform and ad accounts, triggers data
> ingestion, and sees unified metrics (ROAS, CAC) in a single dashboard. The system
> works when a user can go from zero to seeing their first unified metrics panel
> in under 10 minutes.

**If the user cannot articulate this:** they are not ready for domain modeling.
Say so directly and help them clarify the product first.

**Wait for user validation before proceeding.**

## Step 1: Extract business concepts

**Goal:** Identify the raw material for the domain model.

From the requirements and use cases, extract systematically:

1. **Every noun** → candidate entities and value objects.
   (collaborator, discount, store, redemption, ingestion run)
2. **Every verb** → candidate commands and operations.
   (redeem, cancel, configure, ingest, activate)
3. **Every constraint** ("must", "cannot", "only if", "at most", "never", "always")
   → candidate invariants.
   ("cannot exceed 3 uses per day", "must be active to redeem")
4. **Every external system** → integration boundaries, potential context borders.
   (POS system, Meta Ads API, DynamoDB legacy table)
5. **Every ambiguous term** — same word meaning different things to different people
   → bounded context boundaries.
   ("account" = user account vs. ad account vs. billing account)

Present the extracted lists to the user. Ask:
- "Did I miss any important concepts?"
- "Are any of these not actually relevant to the system?"
- "Do any of these terms mean different things in different parts of the business?"

**Wait for user validation before proceeding.**

## Step 2: Define bounded contexts

**Goal:** Divide the domain into autonomous subsystems with unambiguous language.

For each bounded context, document:
- **Name**: clear and descriptive
- **Responsibility**: what this context owns (one sentence)
- **Why it is separate**: what makes it distinct
- **Concepts that live here**: entities, value objects, processes
- **Ambiguities resolved**: terms clarified within this boundary
- **Priority**: core, supporting, or generic

**Context priority determines modeling depth:**
- **Core**: business differentiator. Model in full depth. This is where invariants,
  state machines, and domain events matter most.
- **Supporting**: necessary for core to work, not the differentiator. Solid entity
  definitions and key invariants.
- **Generic**: commoditized functionality (auth, notifications, audit). Minimal
  modeling — reference the external system or library.

**Include a relationship diagram** showing how contexts interact: which feeds data
to which, which emits events that others consume. Use a simple text format:

```
[Data Ingestion] --CollaboratorDeactivated--> [Redemptions]
[Redemptions] --RedemptionConfirmed--> [Audit]
[Discounts] <--reads configuration-- [Redemptions]
```

**Wait for user validation before proceeding.**

## Step 3: Extract entities, aggregates, and value objects

**Goal:** Define what exists in the business, what identity means, and who guards
consistency.

For each bounded context (depth based on priority), identify:

- **Entities**: things with identity and lifecycle. It matters *which* one.
- **Value objects**: defined by attributes, no identity. Two with same attributes
  are interchangeable.
- **Aggregate roots**: entities that guard consistency boundaries. All modifications
  go through the root.

### Entity format

| Attribute | Semantic type | Description | Constraints |
|-----------|--------------|-------------|-------------|
| (name) | (type) | (meaning) | (rules) |

**Always use semantic types, never database types:**

| Semantic type | Meaning | Examples |
|---------------|---------|----------|
| identifier | Unique identity for this entity | collaborator ID, redemption ID |
| natural-key | Identity from external system, already standardized | store code "2104", brand code "BB" |
| text | Free-form text | name, description, address |
| quantity | Countable number (integer, >= 0) | daily uses, processed records |
| money | Monetary amount with precision | sale amount, discount amount |
| percentage | Value 0-100 | discount percentage |
| date | Calendar date, no time | birth date, validity start |
| moment | Point in time | created at, confirmed at |
| time-of-day | Time without date | scheduled execution hour |
| duration | Time span | birthday window in days |
| flag | Yes/no | is aggregator, active |
| state | Finite named states | ACTIVE/INACTIVE, PENDING/CONFIRMED |
| category | Finite named categories | DNI/CE/Passport, NGR/INTERCORP |
| reference | Relationship to another entity | "belongs to company X" |
| flexible-structure | Semi-structured data | error details, audit snapshots |

### Value object format

```
Document (Value Object)
  - number: text
  - type: category (DNI, CE, Passport)
  - Equality: same number + same type = same document
  - Rules: number is non-empty, type is one of the known categories
```

### Guidelines
- References between entities use identity (identifier or natural-key), never
  mutable attributes like names.
- Document nullable attributes with rationale.
- For catalogs (small, known sets), list known values.
- Flag open problems explicitly with recommendations.

**Present entities context by context. Wait for user validation before proceeding.**

## Step 4: Define invariants

**Goal:** Make business rules explicit, unbreakable, and traceable.

Invariants are rules that **cannot be broken** regardless of entry channel
(API, web, CLI, cron, message queue).

For each invariant:
- **ID**: INV-01, INV-02, ...
- **Rule**: clear, unambiguous sentence starting with a subject
- **Applies to**: which operations/services
- **Example**: concrete case, especially edge cases

**Group by category:**
- Core business rules (status checks, limits, ownership)
- Temporal rules (date windows, same-day, timezone)
- Concurrency rules (mutual exclusion, atomic operations)
- Audit rules (mandatory logging, immutability)
- Data integrity rules (uniqueness, non-negativity, referential integrity)

**Produce a validation matrix** — table with operations as columns, invariants as
rows, marked with `x` where each applies.

**Example:**

```markdown
## INV-03 — Daily usage limit per brand

A collaborator cannot exceed the maximum daily uses configured for a given discount
type and brand. Evaluated per calendar day in America/Lima timezone.

**Applies to:** Reserve Redemption, Confirm Redemption
**Example:** Config allows 2 uses/day for Bembos. Collaborator redeemed twice today
for Bembos. Third attempt is rejected — even if Papa John's quota is unused.
```

**Wait for user validation before proceeding.**

## Step 5: Identify domain events

**Goal:** Define how bounded contexts communicate.

For each state transition or operation with cross-context impact:

| Event | Trigger | Source context | Consumers | Payload |
|-------|---------|---------------|-----------|---------|
| (past-tense name) | (what causes it) | (emitter) | (listeners) | (key data) |

**Rules for identifying events:**
- State transition in one context requires action in another → domain event
- Operation needs to notify external systems → domain event
- Audit trail of "what happened" → domain event
- Name in past tense: RedemptionConfirmed, CollaboratorDeactivated

**Tell the user:** these events map directly to technical infrastructure later
(EventBridge, SNS, SQS, webhooks). Defining them now means the event-driven
architecture writes itself.

**Wait for user validation before proceeding.**

## Step 6: Map state machine lifecycles

**Goal:** Document how stateful entities change over time.

For every entity with a `state` attribute:

1. **ASCII state diagram:**
```
    ┌──────────┐     POS confirms     ┌─────────────┐
    │ PENDING  │ ──────────────────▶  │  CONFIRMED  │
    └──────────┘                      └─────────────┘
         │                                  │
         │ timeout / cancel                 │ void
         ▼                                  ▼
    ┌──────────┐                      ┌──────────┐
    │CANCELLED │                      │  VOIDED  │
    └──────────┘                      └──────────┘
```

2. **Valid transitions table:**

| From | To | Trigger | Side effects | Event emitted |
|------|----|---------|-------------|---------------|
| PENDING | CONFIRMED | POS confirmation | — | RedemptionConfirmed |
| PENDING | CANCELLED | Timeout or user | Counter decremented | RedemptionCancelled |

3. **Invalid transitions** — list explicitly what cannot happen and why.

4. **Temporal semantics** — what "same day" means, which timezone governs.

**Wait for user validation before proceeding.**

## When to stop modeling

Tell the user the model is ready when:

1. Every use case can be described as a command that touches one aggregate and
   respects known invariants.
2. Every invariant traces to at least one entity and one operation.
3. Every bounded context has at least one entity and a clear responsibility.
4. Context communication is defined through events.

If these criteria are met, stop. Do not model concepts without a use case driving
them. Do not add attributes "just in case." Perfectionism in modeling is the same
trap as premature optimization in code.

## Producing the final documents

Once all steps are validated, produce the output files.

**Output directory:** `analysis/domain-model/`

### Files to produce:

**`bounded-contexts.md`**
- Value journey (Step 0)
- Each bounded context: responsibility, rationale, concepts, ambiguities, priority
- Relationship diagram with event flows

**`entities-aggregates.md`**
- Timezone and domain conventions
- Per context: entities (semantic types), value objects, business rules,
  relationships, aggregate roots
- Open problems and recommendations

**`invariants.md`**
- Brief definition of invariants
- Each invariant: ID, rule, scope, examples
- Grouped by category
- Validation matrix

**`domain-events.md`**
- Event catalog: name, trigger, source, consumers, payload
- Event flow diagram

**`lifecycle.md`**
- State diagrams (ASCII) per stateful entity
- Valid transitions with triggers, side effects, events
- Invalid transitions
- Temporal semantics

## Iterating the model

The domain model is a living document. Tell the user:

- **New invariant discovered during implementation?** Add it to the model before
  coding around it.
- **Use case changed?** Update affected entities and invariants.
- **Two contexts should be one (or vice versa)?** Refactor the model — cheaper
  than refactoring code.
- **After first production release?** Revisit. Reality teaches what requirements cannot.

## Checklist

Before producing final documents, verify:

- [ ] Value journey is defined and anchors the model
- [ ] Every bounded context has responsibility, rationale, and priority
- [ ] Ambiguous terms are resolved per context
- [ ] Every entity uses **semantic types** (no database types anywhere)
- [ ] Value objects are identified with composition and equality rules
- [ ] Aggregate roots have defined consistency boundaries
- [ ] References use stable identifiers, not mutable names
- [ ] All invariants are numbered, scoped, and exampled
- [ ] Validation matrix covers all operations
- [ ] Domain events are cataloged with source, consumers, and payload
- [ ] Every stateful entity has a lifecycle diagram
- [ ] Invalid transitions are documented
- [ ] Temporal semantics are explicit
- [ ] Open problems are flagged with recommendations
- [ ] Every element traces to a use case and the value journey
- [ ] "When to stop" criteria are met
