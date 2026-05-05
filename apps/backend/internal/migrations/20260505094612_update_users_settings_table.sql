-- Rename a column from "gemini_api_key_encrypted" to "api_key_encrypted"
ALTER TABLE "public"."user_settings" RENAME COLUMN "gemini_api_key_encrypted" TO "api_key_encrypted";
