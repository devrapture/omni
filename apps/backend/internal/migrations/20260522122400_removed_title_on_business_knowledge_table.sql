-- Drop index "idx_business_knowledges_source_chunk_unique" from table: "business_knowledges"
DROP INDEX "public"."idx_business_knowledges_source_chunk_unique";
-- Modify "business_knowledges" table
ALTER TABLE "public"."business_knowledges" DROP COLUMN "title";
