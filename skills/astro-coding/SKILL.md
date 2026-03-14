---
name: astro-coding
description: Provides Astro/Starlight implementation patterns, best practices, and critical rules with tiered context loading
---

# astro-coding Skill

Smart, context-aware implementation patterns for Astro and Starlight development. Provides the knowledge needed to write high-quality, standards-compliant Astro code.

## Purpose

The astro-coding skill is a **knowledge provider** that supplies Astro-specific coding patterns, critical rules, and best practices. It uses a tiered loading strategy to minimize token usage while ensuring all critical rules are always available.

## Tiered Loading Strategy

### Tier 1: Critical Rules (ALWAYS LOADED - ~100 tokens)
The non-negotiable rules that prevent breaking errors. Loaded for every Astro/Starlight task.

**Source**: `./references/critical-rules.md`

**Contains**:
1. File extensions in imports (`.astro`, `.ts`, `.js`)
2. Correct module prefixes (`astro:content` not `astro/content`)
3. Use `class` not `className` in .astro files
4. Async operations in frontmatter only
5. Never expose `SECRET_*` client-side
6. Type all component Props interfaces
7. Define `getStaticPaths()` for dynamic routes
8. Don't access `Astro.params` inside `getStaticPaths()`
9. Use proper collection types (`CollectionEntry<'name'>`)
10. Validate XSS risk with `set:html`

### Tier 2: Common Patterns (CONTEXT-LOADED - ~400 tokens)
Pattern-specific knowledge loaded based on task type detection.

**Sources**:
- `./references/astro-patterns.md` - Core Astro patterns
- `./references/error-catalog.md` - 100+ error patterns indexed by symptom

**Load based on keywords**:

- **"component"** → Component patterns, TypeScript patterns
- **"page" or "route"** → Routing patterns, dynamic routes, `getStaticPaths`
- **"collection" or "content"** → Content collections, schemas, queries
- **"config" or "integration"** → Configuration patterns
- **"starlight"** → Starlight patterns, sidebar, components
- **"error" or "fix" or "debug"** → Error catalog for diagnostic help

## Critical Rules Reference

**These rules are ALWAYS enforced, loaded from Tier 1:**

### 1. File Extensions Required ✅
```typescript
// ✅ CORRECT
import Header from './Header.astro';
import { formatDate } from '../utils/dates.ts';

// ❌ WRONG - Build error
import Header from './Header';
```

### 2. Correct Module Prefixes ✅
```typescript
// ✅ CORRECT - Use colon
import { getCollection } from 'astro:content';

// ❌ WRONG - Module not found
import { getCollection } from 'astro/content';
```

### 3. Use `class` Not `className` ✅
```astro
<!-- ✅ CORRECT in .astro files -->
<div class="container">

<!-- ❌ WRONG - React/JSX syntax -->
<div className="container">
```

### 4. Await in Frontmatter Only ✅
```astro
---
// ✅ CORRECT
const posts = await getCollection('blog');
---
<ul>{posts.map(p => <li>{p.data.title}</li>)}</ul>

<!-- ❌ WRONG - Await in template -->
<ul>{(await getCollection('blog')).map(...)}</ul>
```

### 5. Never Expose Secrets ✅
```typescript
// ✅ CORRECT - Server-side only
const apiKey = import.meta.env.SECRET_API_KEY;

// ❌ WRONG - Exposed to client
<script>
  const key = import.meta.env.SECRET_API_KEY; // ❌ NEVER
</script>
```

## Quick Reference Templates

### Basic Component
```astro
---
interface Props {
  title: string;
  items: string[];
  variant?: 'primary' | 'secondary';
}

const { title, items, variant = 'primary' } = Astro.props;
---

<div class={`component component--${variant}`}>
  <h2>{title}</h2>
  <ul>
    {items.map(item => <li>{item}</li>)}
  </ul>
</div>

<style>
  .component {
    padding: 1rem;
  }
  .component--primary {
    background: var(--color-primary);
  }
</style>
```

### Dynamic Route
```astro
---
// File: src/pages/blog/[slug].astro
import { getCollection } from 'astro:content';
import type { CollectionEntry } from 'astro:content';

export async function getStaticPaths() {
  const posts = await getCollection('blog');
  return posts.map(post => ({
    params: { slug: post.slug },
    props: { post },
  }));
}

interface Props {
  post: CollectionEntry<'blog'>;
}

const { post } = Astro.props;
const { Content } = await post.render();
---

<article>
  <h1>{post.data.title}</h1>
  <time datetime={post.data.publishDate.toISOString()}>
    {post.data.publishDate.toLocaleDateString()}
  </time>
  <Content />
</article>
```

### Content Collection
```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content';

const blog = defineCollection({
  type: 'content',
  schema: z.object({
    title: z.string(),
    description: z.string(),
    publishDate: z.date(),
    author: z.string(),
    tags: z.array(z.string()).optional(),
    draft: z.boolean().default(false),
  }),
});

export const collections = { blog };
```

## Task-Based Loading Examples

### Simple Component Task
```
Task: "Create a Card component"
Load:
  - Tier 1: Critical rules (always)
  - Tier 2: Component patterns from astro-patterns.md
Tokens: ~300 total
```

### Route with Collections
```
Task: "Add blog with pagination"
Load:
  - Tier 1: Critical rules (always)
  - Tier 2: Routing patterns + Collection patterns
Tokens: ~600 total
```

### Complex Integration
```
Task: "Integrate GitBook API with custom loader"
Load:
  - Tier 1: Critical rules (always)
  - Tier 2: Collection patterns + error catalog
Tokens: ~1200 total
```

### Bug Fix
```
Task: "Fix TypeScript errors in components"
Load:
  - Tier 1: Critical rules (always)
  - Tier 2: Error catalog + TypeScript patterns
Tokens: ~400 total
```

## Pattern Detection Keywords

The skill detects keywords to load appropriate patterns:

| Keywords | Patterns Loaded |
|----------|-----------------|
| component, card, button, layout | Component patterns, Props typing |
| page, route, [slug], dynamic | Routing patterns, getStaticPaths |
| collection, content, blog, docs | Collection patterns, schemas |
| config, integration, tailwind | Configuration patterns |
| error, fix, debug, failing | Error catalog |
| authentication, auth, login | Security patterns + integrations |

## Token Optimization Guidelines

**Minimal loading** (~100 tokens):
- Only critical rules
- For trivial tasks (<10 lines, 1 file)
- Example: Fix typo, update text

**Standard loading** (~400 tokens):
- Critical rules + relevant pattern section
- For typical tasks (20-100 lines, 2-5 files)
- Example: Create component, add route

**Full loading** (~1200 tokens):
- Critical rules + multiple patterns
- For complex tasks (>100 lines, >5 files, integrations)
- Example: Custom loaders, multi-source systems, refactoring

## Error Catalog Usage

The error catalog indexes 100+ error patterns by symptom for quick diagnosis:

**Example lookup**:
```
Symptom: "Cannot find module './Header'"
→ Error catalog points to: Missing file extension
→ Solution: Add .astro extension to import
```

**Categories in catalog**:
- Import errors
- Component errors
- Routing errors
- Collection errors
- Configuration errors
- TypeScript errors
- Runtime errors

## Best Practices

**When writing code with this skill**:
1. ✅ Always check critical rules first
2. ✅ Load only relevant pattern sections
3. ✅ Use error catalog for debugging
4. ✅ Type all Props interfaces
5. ✅ Include error handling for async operations
6. ✅ Consider accessibility (ARIA labels, semantic HTML)
7. ✅ Optimize performance (client directives, image handling)

**Validation checklist** (from critical rules):
- [ ] All imports have file extensions
- [ ] Using `astro:content` not `astro/content`
- [ ] Using `class` not `className`
- [ ] All `await` in frontmatter, not templates
- [ ] No `SECRET_*` in `<script>` tags
- [ ] All Props have TypeScript interfaces
- [ ] Dynamic routes have `getStaticPaths()`

## Version

**Skill Version**: 1.0 
**Last Updated**: 2026-03-13

This skill provides comprehensive Astro/Starlight knowledge with intelligent, token-efficient loading based on task requirements.

