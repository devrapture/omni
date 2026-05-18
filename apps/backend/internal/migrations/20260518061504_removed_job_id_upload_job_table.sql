-- atlas:txtar

-- migration.sql --
-- atlas:nolint DS103
-- Modify "upload_jobs" table
ALTER TABLE "public"."upload_jobs" DROP COLUMN "job_id";