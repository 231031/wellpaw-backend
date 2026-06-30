# API Update — Pet Food Plan Total decoupled from Pet Detail (2026-06-29)

## What changed (one line)

`pet_detail_id` (and the nested `pet_detail` object) **moves off** every `pet_food_plan_total`
and **onto** each `pet_food_plan_history`.

## Why

Previously the plan **total** was wired to a specific pet-detail snapshot, so *every* pet-detail
update created a new total — even for owner-defined (`self_define`) plans where the feeding amounts
should stay the same. Now the pet-detail ↔ total association lives on the **history** row (the
point-in-time record). As a result:

- For a `self_define` plan, changing the pet detail (weight, BCS, activity, …) **no longer changes
  the plan total**. Only a new history row is recorded.
- For an app-calculated plan, behaviour is unchanged: amounts are recalculated and a new total is
  produced.

## Request bodies

**No changes.** No request ever sent `pet_detail_id` on a total, so nothing on the Flutter request
side needs to change.

## Response changes

### 1. `pet_food_plan_total` object — REMOVED fields

The following fields are **removed** from every `pet_food_plan_total` wherever it appears:

- `pet_detail_id`
- `pet_detail` (nested object — was only populated on create/calculate)

**Affected endpoints / paths:**

| Endpoint | JSON path that lost the fields |
|---|---|
| `POST /foodplan/calculate` | `data.pet_food_plan.pet_food_plan_totals[].pet_detail_id`, `.pet_detail` |
| `POST /foodplan/` (create) | `data.pet_food_plan.pet_food_plan_totals[].pet_detail_id`, `.pet_detail` |
| `GET /foodplan/:pet_id` | `data.pet_food_plan.pet_food_plan_totals[].pet_detail_id` |
| `POST /pet/detail` | `data.pet_food_plan.pet_food_plan_totals[].pet_detail_id` |
| `GET /pet/analysis/:pet_id` | `data.pet_food_plan_histories[].pet_food_plan_total.pet_detail_id` |

**Before:**
```json
"pet_food_plan_total": {
  "id": 42,
  "pet_food_plan_id": 7,
  "pet_detail_id": 31,
  "total_energy_intake": 350.0,
  "total_protein_intake": 22.0,
  "total_fat_intake": 9.5,
  "created_at": "2026-06-29T10:00:00+07:00",
  "pet_detail": { "...": "..." }
}
```

**After:**
```json
"pet_food_plan_total": {
  "id": 42,
  "pet_food_plan_id": 7,
  "total_energy_intake": 350.0,
  "total_protein_intake": 22.0,
  "total_fat_intake": 9.5,
  "created_at": "2026-06-29T10:00:00+07:00"
}
```

### 2. `pet_food_plan_history` object — ADDED fields

Each history item now carries the pet-detail snapshot that was active for that record:

- `pet_detail_id` (uint)
- `pet_detail` (nested object — full pet detail snapshot)

**Affected endpoint:** `GET /pet/analysis/:pet_id` → `data.pet_food_plan_histories[]`

**Before:**
```json
{
  "id": 100,
  "pet_id": 3,
  "pet_food_plan_total_id": 42,
  "created_at": "2026-06-29T10:00:00+07:00",
  "plan_usage_end_date": "2026-06-29T12:00:00+07:00",
  "pet_food_plan_total": { "...": "..." }
}
```

**After:**
```json
{
  "id": 100,
  "pet_id": 3,
  "pet_food_plan_total_id": 42,
  "pet_detail_id": 31,
  "created_at": "2026-06-29T10:00:00+07:00",
  "plan_usage_end_date": "2026-06-29T12:00:00+07:00",
  "pet_food_plan_total": { "...": "..." },
  "pet_detail": {
    "id": 31,
    "pet_id": 3,
    "weight": 4.2,
    "activity_level": 1,
    "age_range": 1,
    "bcs": 5,
    "lactation": false,
    "gestation": false,
    "neutered": true,
    "energy": 220.0,
    "protein": 14.3,
    "fat": 4.9,
    "created_at": "2026-06-29T10:00:00+07:00"
  }
}
```

## Action for Flutter

- If any screen reads `pet_detail_id` / `pet_detail` from a **plan total**, move it to read from the
  **history** object instead.
- The plan analysis screen can now show, per history entry, the pet detail snapshot that was active
  at that time via `pet_food_plan_histories[].pet_detail`.
- No request payloads change.

> The generated swagger files (`doc/docs.go`, `doc/swagger.json`, `doc/swagger.yaml`) should be
> regenerated (`swag init`) so the published contract matches; this document is the authoritative
> summary of the field changes.
