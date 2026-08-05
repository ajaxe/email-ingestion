# Architecture Decision Records

This directory contains the Architecture Decision Records (ADRs) for the **Email Ingestion Gateway**. ADRs capture significant architectural and design decisions, the context that motivated them, the alternatives considered, and the consequences of the chosen approach.

> **New ADR?** Copy the template below, assign the next sequential number, and submit it for review. An ADR should be written at the time the decision is made — not retroactively reconstructed.

---

## ADR Template

```markdown
# ADR NNNN: <Short imperative title>

## Status

<See Status Vocabulary below>

## Context

<What situation, requirement, or constraint forced this decision?
Include relevant constraints, forces, and prior art considered.>

## Decision

<What was decided? State it clearly and in the active voice.
Include schema snippets, pseudocode, or configuration where it
aids understanding.>

## Consequences

### Positive

* <Benefit A>
* <Benefit B>

### Negative

* <Drawback or trade-off A>
* <Drawback or trade-off B>
```

---

## Status Vocabulary

Every ADR must carry exactly one of the following statuses. The status reflects the current standing of the decision — it is not a historical artifact.

| Status | Meaning |
|---|---|
| **Proposed** | The decision is under active discussion and has not yet been ratified. The described approach may change. |
| **Accepted** | The decision has been ratified and is the current authoritative approach. Implementation may or may not be complete. |
| **Deprecated** | The decision was once accepted but is no longer recommended. It has not yet been formally replaced. New work should avoid this pattern. |
| **Superseded by ADR-NNNN** | The decision has been formally replaced by a newer ADR. The old record is retained for historical context. Link to the superseding ADR. |
| **Rejected** | The decision was proposed but not accepted. The record is retained to document why the approach was ruled out, preventing it from being re-proposed without new context. |
| **Deferred** | The decision has been identified as necessary but is explicitly postponed. The ADR documents the known options and the reason for deferral. |

---

## Decision Register

| ADR | Title | Status | Date |
|---|---|---|---|
| [0001](./0001-smtp-edge-proxy-architecture.md) | Decoupling SMTP Daemon as a Stateless Edge Proxy | Accepted | — |
| [0002](./0002-dual-principal-api-model.md) | Dual Principal Model — Separate API Surfaces for M2M and Human Actors | Accepted | 2026-08-03 |
| [0003](./0003-local-identity-registry.md) | Local Identity Registry Over Custom IdP Claims | Accepted (Partially Superseded) | 2026-08-03 |
| [0004](./0004-application-id-url-placement.md) | `application_id` Placement in API URLs | Accepted | 2026-08-03 |
| [0005](./0005-jit-user-provisioning.md) | Just-in-Time User Provisioning via Email Correlation | Accepted | 2026-08-03 |
| [0006](./0006-organization-grouping-concept.md) | Organization as a Thin Grouping Concept for Future Multi-User Expansion | Accepted | 2026-08-03 |
| [0007](./0007-api-key-m2m-authentication.md) | Gateway-Managed API Keys for M2M Authentication | Accepted | 2026-08-04 |

---

## Relationships Between ADRs

Some decisions build directly on others. The dependency graph below shows which ADRs must be understood together.

```
0001  SMTP Edge Proxy
  │
  └── (independent foundation)

0002  Dual Principal Model (M2M vs Human)
  │
  ├── 0003  Local Identity Registry          (how human principals are resolved)
  │     │
  │     ├── 0005  JIT User Provisioning      (how human principals are activated)
  │     │     │
  │     │     └── 0006  Organization Concept (grouping added to activation flow)
  │     │
  │     └── 0007  API Key M2M Authentication (supersedes 0003 for M2M identity resolution)
  │
  └── 0004  application_id URL Placement     (consequence of two principal types)
```

---

## Governance

- ADRs are **immutable once Accepted**. Corrections to factual errors may be made as minor edits with a note at the top of the document. Design changes require a new ADR that supersedes the old one.
- The **Status** field must be updated when a decision is superseded or deprecated — do not leave stale `Accepted` statuses on replaced decisions.
- ADRs are numbered sequentially. Gaps in the sequence indicate rejected or withdrawn proposals; do not re-use numbers.
- When a decision is superseded, add `Superseded by ADR-NNNN` to the old record's Status field and link back from the new one.
