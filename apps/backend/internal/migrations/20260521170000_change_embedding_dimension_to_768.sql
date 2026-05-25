-- Modify "business_knowledges" table
ALTER TABLE "public"."business_knowledges"
  ALTER COLUMN "embedding" TYPE public.vector(768)
  USING CASE
    WHEN "embedding" IS NULL OR vector_dims("embedding") = 768 THEN "embedding"::public.vector(768)
    ELSE NULL
  END;
