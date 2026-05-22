-- Remove duplicate active chunks before adding the uniqueness guard.
WITH ranked AS (
  SELECT
    ctid,
    row_number() OVER (
      PARTITION BY "business_id", "source_name", "chunk_index"
      ORDER BY "created_at" DESC NULLS LAST, ctid DESC
    ) AS row_num
  FROM "public"."business_knowledges"
  WHERE "deleted_at" IS NULL
)
DELETE FROM "public"."business_knowledges"
WHERE ctid IN (
  SELECT ctid
  FROM ranked
  WHERE row_num > 1
);

-- Create index "idx_business_knowledges_source_chunk_unique" to prevent duplicate active chunks.
CREATE UNIQUE INDEX "idx_business_knowledges_source_chunk_unique"
ON "public"."business_knowledges" ("business_id", "source_name", "chunk_index")
WHERE "deleted_at" IS NULL;
