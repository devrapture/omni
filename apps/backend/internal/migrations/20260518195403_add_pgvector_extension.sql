-- Create extension "vector"
CREATE EXTENSION IF NOT EXISTS "vector" WITH SCHEMA "public" VERSION "0.8.2";
-- Modify "business_knowledges" table
ALTER TABLE "public"."business_knowledges" DROP CONSTRAINT "fk_business_knowledges_user", ADD COLUMN "embedding" public.vector(1536) NULL, ADD COLUMN "embedding_model" text NOT NULL DEFAULT 'gemini-embedding-001', ADD CONSTRAINT "fk_users_business_knowledge" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
