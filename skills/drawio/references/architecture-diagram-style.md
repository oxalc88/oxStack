# Architecture diagram style (native AWS4 shapes)

A polished-diagram recipe distilled from real corrections on `ngr-0000-orders-migration`'s trigger/runtime diagram, cross-checked against a hand-built reference (`ngr-0000-api/symphony/Product-Definition/architecture/diagrams/symphony-architecture.drawio.xml`) and the actual `mxgraph.aws4` stencil bundled in the draw.io desktop app. Follow this whenever the request is an AWS architecture diagram — it prevents the multi-round "spacing is wrong / icon is wrong / this crosses funny" correction cycle.

## 1. Verify `resIcon` names before using them — don't guess

A wrong `resIcon` string doesn't error, it silently renders a generic/wrong-looking icon (e.g. `mxgraph.aws4.registry` renders as a generic grid icon, not ECR — the real name is `mxgraph.aws4.ecr`). If the draw.io desktop app is installed, confirm the exact shape name against its bundled stencil instead of guessing from memory:

```bash
mkdir -p /tmp/drawio_stencils && cd /tmp/drawio_stencils
npx --yes @electron/asar extract "/Applications/draw.io.app/Contents/Resources/app.asar" extracted
grep -n '<shape aspect="fixed"[^>]*name="ecr"' extracted/drawio/src/main/webapp/stencils/aws4.xml
```

Every `<shape name="some name">` in `stencils/aws4.xml` (inside the `<shapes name="mxgraph.aws4">` wrapper) is reachable as `resIcon=mxgraph.aws4.some_name` (spaces become underscores). Grep for the service name (`ecr`, `lambda`, `dynamodb`, `eventbridge`, `sns`, `cloudwatch logs`, `fargate`, `sqs`, `api_gateway`, `client`) before committing to a style string. This is a one-time lookup per unfamiliar service — cheap insurance against a wrong icon shipping in a reviewed document.

## 2. Exact style recipe (confirmed against a real hand-built reference)

Icon size **78x78** (not smaller — 55-60px reads as cramped next to 2-3 line labels). Every resource icon uses this shape:

```
shape=mxgraph.aws4.resourceIcon;resIcon=mxgraph.aws4.<service>;
outlineConnect=0;fontColor=#232F3E;strokeColor=#ffffff;dashed=0;
verticalLabelPosition=bottom;verticalAlign=top;align=center;html=1;
fontSize=11;fontStyle=0;aspect=fixed;
gradientDirection=north;gradientColor=<family-light>;fillColor=<family-dark>;
```

`strokeColor=#ffffff` (a thin white outline) is what separates the icon cleanly from the page background — omitting it makes icons look flat/blend together.

Confirmed family colors (fillColor / gradientColor):

| Family | fillColor | gradientColor | Services |
|---|---|---|---|
| Compute | `#D05C17` | `#F78E04` | `lambda`, `fargate`, `ecr` |
| Messaging / eventing | `#BC1356` | `#FF4F8B` | `sqs`, `sns`, `eventbridge` |
| Management/Governance | `#E7157B` | `#FF4F8B` | `cloudwatch_logs`, `cloudwatch` |
| Database | `#3334B9` | `#4D72F3` | `dynamodb` |
| Networking | `#5A30B5` | `#945DF2` | `api_gateway` |
| Storage | `#277116` | `#60A337` | `s3` |
| Generic actor (non-AWS) | `#232F3E` | none | `client` (people, POS terminals, external systems) |

For a non-AWS external system/actor (a third-party aggregator, a POS device, a person), don't force an AWS icon onto it — use either `shape=mxgraph.aws4.client` (flat dark icon, no gradient) or a plain rounded box: `rounded=1;whiteSpace=wrap;html=1;fillColor=#fbebdc;strokeColor=#c96a1f;fontSize=12;`.

## 3. Section grouping: plain dashed rectangles, not swimlanes

Don't use `swimlane` with `startSize` for section containers ("AWS Account", "External System X") — the title bar has a fixed height and the title text truncates silently if the box is narrower than the text needs (you won't see an error, just a cut-off label). Use a plain rectangle instead, with the label wrapping at the top:

```
fillColor=none;dashed=1;verticalAlign=top;fontStyle=1;fontColor=<accent>;
strokeColor=<accent>;fontSize=13;strokeWidth=2;whiteSpace=wrap;html=1;spacingBottom=10;
```

Give it generous width/height so the label wraps across 1-2 lines without crowding the first row of icons (leave at least 80-100px between the label and the first icon row). Icons inside the box are **not** parented to it — give every icon `parent="1"` with absolute coordinates that happen to fall inside the box's rectangle. This avoids container-clipping/z-order surprises and matches how real hand-built diagrams in this codebase do it.

## 4. Spacing: generous grid, no exceptions

- Row pitch (vertical distance between rows of icons): **160-200px minimum**.
- Column pitch (horizontal distance between icon centers in the same row): **200-260px minimum**.
- Section box padding: at least 60-80px between the box border and the icons/labels inside it.
- Never rely on default/automatic label placement to save you from tight spacing — if two nodes are close enough that their labels *could* touch, they're too close.

## 5. Fan-out from one node: stagger the elbow height, don't bunch

The most common failure mode: one node (e.g. a compute task) has 4+ outgoing edges to targets spread out in the same row below it. Left to default routing, all 4 edges exit the source within a tiny span (a 78px-wide icon only gives ~15px between four evenly-spaced exit points), then run their horizontal label segments at the same height — labels collide.

Fix: give each edge its own explicit waypoint at a **different y** (a "staircase"), so each edge's horizontal run — and its label — sits on its own row:

```xml
<mxCell id="e1" edge="1" parent="1" source="task" target="target1"
  style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;
  exitX=0.2;exitY=1;entryX=0.5;entryY=0;fontSize=11;labelBackgroundColor=#ffffff;" value="label 1">
  <mxGeometry relative="1" as="geometry">
    <Array as="points"><mxPoint x="TARGET1_CENTER_X" y="400" /></Array>
  </mxGeometry>
</mxCell>
<mxCell id="e2" ... exitX="0.4" ...>
  <mxGeometry relative="1" as="geometry">
    <Array as="points"><mxPoint x="TARGET2_CENTER_X" y="450" /></Array>
  </mxGeometry>
</mxCell>
```

Increment the waypoint `y` by ~40-50px per additional edge, and spread `exitX` evenly (0.2/0.4/0.6/0.8 for four edges). Always add `labelBackgroundColor=#ffffff` to edges that might run near another element — it makes the label legible even if a line passes close by.

## 6. Separate build-time/deploy-time resources from the runtime flow

A container registry (ECR), an artifact bucket, a CI pipeline — anything that's only touched at build/deploy time — does not belong inline in a runtime/trigger sequence diagram. Drawing it as if "the task pulls the image" is a live step in the same flow as the trigger chain reads as more architecturally significant than it is, and it clutters the primary flow with a mostly-static relationship. Give it its own clearly-labeled, visually distinct box (e.g. `fontStyle=2` italic, muted gray `#93a1ab` border/text, label like "IMAGE REGISTRY -- separate stack (path/to/template.yaml), deploy-time only"), physically separated from the main flow, with a muted dashed connector.

## 7. Don't assert platform behavior you haven't verified

A diagram label is a factual claim. "Image pulled on every task start" asserts something specific about a cloud platform's caching behavior that changes over AWS releases and isn't verifiable from this repo's own config. Prefer the plain, defensible relationship: "image pull" (ECS pulls this image to run the task) — true regardless of caching internals. The same discipline that applies to prose docs (state only what's evidenced by code/config, don't editorialize) applies to every edge label in a diagram.

## 8. Always render and visually verify before calling it done

Never ship a hand-authored `.drawio` file's layout on faith. If the draw.io CLI is available:

```bash
"/Applications/draw.io.app/Contents/MacOS/draw.io" -x -f png -e -b 20 -o preview.png diagram.drawio
```

Read the PNG back. For dense regions (a fan-out, a busy corner), crop and re-inspect at higher resolution before trusting it's readable:

```bash
python3 -c "
from PIL import Image
im = Image.open('preview.png')
im.crop((x0, y0, x1, y1)).save('preview_crop.png')
"
```

Delete the preview PNG(s) once satisfied — they're a verification step, not a deliverable. Iterate (adjust XML → re-render → re-inspect) until a full-resolution read of the image shows no overlapping labels, no crossed lines through text, and no truncated titles. This costs a few tool calls; skipping it costs a multi-round correction cycle with the user instead.

## 9. Title, subtitle, legend, footer

Match this structure for any non-trivial architecture diagram (all plain `text;html=1;` cells, `parent="1"`):

- **Title**: `fontSize=20;fontStyle=1;fontColor=#1b2226;` — one line, plain hyphens not em-dashes (avoids XML entity noise).
- **Subtitle**: `fontSize=13;fontColor=#55636b;` — one sentence explaining the "why", directly under the title.
- **Legend**: small 14x14 colored squares (`rounded=0;whiteSpace=wrap;html=1;fillColor=<color>;strokeColor=none;`) paired with a text cell, one per color family actually used in the diagram.
- **Footer**: `fontSize=11;fontStyle=2;fontColor=#93a1ab;` (italic, muted) — cite the exact source files/repos the diagram was built from, and the verification date if anything was checked against live infrastructure. Treat this the same as a citation in a written doc.
