-- Create "upload_jobs" table
CREATE TABLE "public"."upload_jobs" (
  "id" uuid NOT NULL,
  "status" text NOT NULL DEFAULT 'queued',
  "user_id" uuid NOT NULL,
  "file_path" text NOT NULL,
  "source_type" text NOT NULL,
  "content" text NULL,
  "error" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_upload_jobs_user_id" to table: "upload_jobs"
CREATE INDEX "idx_upload_jobs_user_id" ON "public"."upload_jobs" ("user_id");
