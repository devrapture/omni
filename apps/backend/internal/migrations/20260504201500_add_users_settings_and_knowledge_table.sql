-- Create "business_knowledges" table
CREATE TABLE "public"."business_knowledges" (
  "id" uuid NOT NULL,
  "title" text NOT NULL,
  "content" text NOT NULL,
  "source_name" text NOT NULL,
  "source_type" text NOT NULL,
  "chunk_index" bigint NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "embedding" text NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_business_knowledges_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "user_settings" table
CREATE TABLE "public"."user_settings" (
  "id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "provider" text NOT NULL DEFAULT 'gemini',
  "mode" text NOT NULL DEFAULT 'platform',
  "gemini_api_key_encrypted" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_user_setting" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
