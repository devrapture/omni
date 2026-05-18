test "migrate" "upload_jobs_job_id_backfill_existing_rows" {
  migrate {
    to = "20260517115629"
  }
  exec {
    sql = <<-SQL
      INSERT INTO "public"."upload_jobs" ("id", "status", "user_id", "object_key", "source_type", "created_at", "updated_at")
      VALUES (
        '11111111-1111-1111-1111-111111111111',
        'queued',
        '22222222-2222-2222-2222-222222222222',
        'uploads/22222222-2222-2222-2222-222222222222/file.pdf',
        '.pdf',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
      );
    SQL
  }
  migrate {
    to = "20260517200751"
  }
  exec {
    sql = "SELECT COUNT(*) FROM \"public\".\"upload_jobs\" WHERE \"job_id\" = \"id\" AND \"job_id\" IS NOT NULL"
    output = "1"
  }
  exec {
    sql = "SELECT COUNT(*) FROM \"public\".\"upload_jobs\" WHERE \"job_id\" IS NULL"
    output = "0"
  }
}

test "migrate" "upload_jobs_job_id_empty_table" {
  migrate {
    to = "20260517115629"
  }
  migrate {
    to = "20260517200751"
  }
  exec {
    sql = "SELECT COUNT(*) FROM \"public\".\"upload_jobs\""
    output = "0"
  }
}
