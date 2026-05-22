-- Create "businesses" table
CREATE TABLE "public"."businesses" (
  "id" uuid NOT NULL,
  "name" text NOT NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_business" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Rename a column from "user_id" to "business_id"
ALTER TABLE "public"."business_knowledges" RENAME COLUMN "user_id" TO "business_id";
-- Modify "business_knowledges" table
ALTER TABLE "public"."business_knowledges" DROP CONSTRAINT "fk_users_business_knowledge", ADD CONSTRAINT "fk_businesses_knowledge" FOREIGN KEY ("business_id") REFERENCES "public"."businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
