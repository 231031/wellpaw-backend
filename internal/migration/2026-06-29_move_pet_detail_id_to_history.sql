-- Migration: move pet_detail_id from pet_food_plan_totals -> pet_food_plan_histories
--
-- Run this ONCE against an existing database. AutoMigrate (migration.go) skips
-- tables that already exist (HasTable), so this ALTER/backfill must be applied
-- manually. Fresh databases get the corrected postgresql_schema.sql directly.
--
-- Order matters: backfill the new column from the current total->detail link
-- BEFORE dropping it from pet_food_plan_totals.

BEGIN;

-- 1. Add the new (nullable for now) column on histories.
ALTER TABLE pet_food_plan_histories ADD COLUMN pet_detail_id INT;

-- 2. Backfill each history row from the pet_detail_id its total currently points at.
UPDATE pet_food_plan_histories h
SET pet_detail_id = t.pet_detail_id
FROM pet_food_plan_totals t
WHERE t.id = h.pet_food_plan_total_id;

-- 3. Lock it down and wire the foreign key.
ALTER TABLE pet_food_plan_histories ALTER COLUMN pet_detail_id SET NOT NULL;
ALTER TABLE pet_food_plan_histories
    ADD CONSTRAINT fk_pet_details_pet_food_plan_histories
    FOREIGN KEY (pet_detail_id) REFERENCES pet_details (id);

-- 4. Drop the now-unused link from totals.
ALTER TABLE pet_food_plan_totals DROP CONSTRAINT IF EXISTS fk_pet_details_pet_food_plan_totals;
ALTER TABLE pet_food_plan_totals DROP COLUMN pet_detail_id;

COMMIT;
