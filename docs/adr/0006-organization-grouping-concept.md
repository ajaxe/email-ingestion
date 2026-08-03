# ADR 0006: Organization as a Thin Grouping Concept for Future Multi-User Expansion

## Status

Accepted (Level 1 — Org as Grouping Record)

## Context

The current access model maps users directly to applications via `user_application_access` (ADR-0005). There is no concept of an organization or team that groups users and applications together.

As the system grows, it is likely that:
- A tenant will want multiple users to manage the same application (e.g., a team sharing dashboard access).
- Role management will need to operate at a group level rather than per-user-per-application.
- Applications may need to be grouped under a shared namespace (e.g., staging vs. production applications within the same company).

Introducing an organization concept later, after the schema is in production, would require a retroactive migration on the `applications` table — a core production table with live data. Adding the scaffolding now, while the schema is young, avoids this cost.

Two levels of implementation were considered. The question was whether to add only the grouping record (Level 1) or also add the membership table that fully replaces the direct user-to-application mapping (Level 2).

## Decision

We will implement **Level 1** — adding `organizations` as a first-class schema concept and linking `applications` to it. No user membership model will be implemented at this stage.

### Level 1 — Accepted: Org as a Grouping Record

Add the `organizations` table and a foreign key on `applications`. During JIT user activation (ADR-0005), a personal organization is automatically created for the user and the application is linked to it.

```sql
CREATE TABLE organizations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    is_personal  BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE applications
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE RESTRICT;
```

The `is_personal` flag distinguishes a user's personal namespace (1:1 user-org, auto-created) from a future team organization (1:N user-org, manually created). This distinction is essential for migrating cleanly to team orgs later without touching the personal org rows.

**Updated JIT activation flow (extends ADR-0005):**

On first login, after validating the JWT and finding the pending `user_application_access` record, the middleware also:

1. Creates a personal organization: `INSERT INTO organizations (name = user_email, is_personal = true)`.
2. Links the application to it: `UPDATE applications SET organization_id = new_org.id WHERE id = app_id AND organization_id IS NULL`.
3. Then proceeds with the existing `ActivateUserAccess` update.

Steps 1–3 and the `ActivateUserAccess` update must execute within a single transaction.

**What this gives us:**
- Every application belongs to an org from day one — no retroactive migration needed later.
- The org concept is established in the data model and visible in tooling and admin queries.
- `user_application_access` continues to be the authorization source of truth for the `UserAuth` middleware.
- No new API surface, no new middleware logic beyond the JIT flow tweak.

---

### Level 2 — Not Accepted (Documented for Future Reference): Full Org Membership Foundation

Level 2 adds an `organization_members` table alongside Level 1, establishing the full schema foundation for team orgs without yet building the feature.

```sql
CREATE TABLE organization_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    idp_user_sub    TEXT,
    role            TEXT NOT NULL DEFAULT 'owner',  -- owner | admin | viewer
    status          TEXT NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    activated_at    TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    UNIQUE (organization_id, email),
    UNIQUE (organization_id, idp_user_sub)
);
```

At JIT activation, in addition to Level 1 steps, the user would be inserted as an `owner` member of their personal org. `user_application_access` would remain active as the interim authorization table, with `organization_members` serving as the future replacement.

**Why Level 2 was deferred:**

- The `organization_members` table mirrors the structure of `user_application_access` almost exactly. Maintaining both simultaneously creates redundancy and requires the `UserAuth` middleware to decide which table is authoritative.
- The data migration from `user_application_access` to `organization_members` is straightforward and low-risk whenever team orgs are actually needed — it does not require touching the `applications` or `organizations` tables.
- The schema cost of Level 2 is low, but the operational clarity cost is not. Until multi-user orgs are a feature, the presence of `organization_members` would be a source of confusion with no active use.

**When Level 2 should be implemented:**

Level 2 becomes the correct next step when any of the following are true:
1. A tenant requests the ability to add a second user to their organization.
2. Role management needs to apply at the organization level rather than per-application.
3. An admin API for org invites is being built.

At that point, the migration is:
```
INSERT INTO organization_members (organization_id, email, idp_user_sub, role, status, activated_at)
SELECT a.organization_id, u.email, u.idp_user_sub, u.role, u.status, u.activated_at
FROM user_application_access u
JOIN applications a ON a.id = u.application_id;
```

After the migration, `UserAuth` middleware switches its authorization lookup from `user_application_access` to `organization_members`, and `user_application_access` is deprecated.

## Consequences

### Positive (Level 1)

- **No retroactive migration:** The `organization_id` FK is added now while the schema is pre-production. Future org features require only new tables, not changes to existing ones.
- **Clean future migration path:** `is_personal` flag makes team org rollout surgical — personal orgs are never touched, only new team orgs use the membership model.
- **Conceptual clarity:** Org is a named, explicit entity in the data model. Admin tooling and observability queries can group by org from day one.
- **Minimal implementation cost:** One new table, one column, and a small extension to the JIT activation transaction.

### Negative (Level 1)

- **`organization_id` nullable during transition:** Until all existing applications (if any) are backfilled, `organization_id` may be `NULL`. Application code must tolerate this. The column should be made `NOT NULL` once all rows are populated.
- **Personal org creation in the hot path:** The JIT activation flow now creates an org and updates `applications` in addition to activating the user. This is a one-time cost per user, but the transaction must be robust and idempotent (`INSERT ... ON CONFLICT DO NOTHING`, `UPDATE ... WHERE organization_id IS NULL`).
