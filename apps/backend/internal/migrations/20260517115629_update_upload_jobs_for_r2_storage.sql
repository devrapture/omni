-- atlas:nolint BC102
-- Rename a column from "file_path" to "object_key"
ALTER TABLE "public"."upload_jobs" RENAME COLUMN "file_path" TO "object_key";
-- atlas:nolint DS103
-- Modify "upload_jobs" table
ALTER TABLE "public"."upload_jobs" DROP COLUMN "content", DROP COLUMN "error";
