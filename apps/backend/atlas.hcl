# apps/backend/atlas.hcl

data "external_schema" "go" {
  program = [
    "go",
    "run",
    "-tags", "atlas",
    "./internal/migrations/loader",
  ]
}

data "composite_schema" "app" {
  schema "public" {
    url = "file://internal/migrations/extensions.hcl"
    schema = schema.public
  }
  schema "public" {
    url = data.external_schema.go.url
  }
}

env "local" {
  src = data.composite_schema.app.url
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
