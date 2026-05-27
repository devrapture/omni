# apps/backend/atlas.hcl

data "external_schema" "go" {
  program = [
    "go",
    "run",
    "-tags", "atlas",
    "./internal/migrations/loader",
  ]
}


env "local" {
  src = data.external_schema.go.url
  url = getenv("DATABASE_URL")
  dev = getenv("ATLAS_DEV_URL")

  migration {
    dir = "file://internal/migrations"
  }

  lint {
    destructive {
      error = false
    }
  }
}
