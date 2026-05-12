-- Modify "business_knowledges" table
ALTER TABLE "public"."business_knowledges" DROP COLUMN "embedding";
-- Create index "idx_user_settings_user_id" to table: "user_settings"
CREATE UNIQUE INDEX "idx_user_settings_user_id" ON "public"."user_settings" ("user_id");
