-- Modify "upload_jobs" table
ALTER TABLE "public"."upload_jobs" ADD COLUMN "source_name" text NULL;

-- Backfill existing rows with the filename portion of the object key.
UPDATE "public"."upload_jobs"
SET "source_name" = regexp_replace("object_key", '^.*/', '')
WHERE "source_name" IS NULL;

ALTER TABLE "public"."upload_jobs" ALTER COLUMN "source_name" SET NOT NULL;
