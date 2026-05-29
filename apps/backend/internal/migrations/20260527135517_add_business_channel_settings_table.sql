-- Create "business_channel_settings" table
CREATE TABLE "public"."business_channel_settings" (
  "id" uuid NOT NULL,
  "business_id" uuid NOT NULL,
  "telegram_bot_token_encrypted" text NULL,
  "telegram_bot_username" text NULL,
  "telegram_active" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_businesses_channel_setting" FOREIGN KEY ("business_id") REFERENCES "public"."businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_business_channel_settings_business_id" to table: "business_channel_settings"
CREATE UNIQUE INDEX "idx_business_channel_settings_business_id" ON "public"."business_channel_settings" ("business_id");
