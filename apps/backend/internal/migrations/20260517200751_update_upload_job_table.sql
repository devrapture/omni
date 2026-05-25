-- Modify "upload_jobs" table
ALTER TABLE "public"."upload_jobs" ADD COLUMN "job_id" uuid NULL;
-- Backfill existing rows
UPDATE "public"."upload_jobs" SET "job_id" = "id" WHERE "job_id" IS NULL;
-- Enforce NOT NULL after backfill
ALTER TABLE "public"."upload_jobs" ALTER COLUMN "job_id" SET NOT NULL;
