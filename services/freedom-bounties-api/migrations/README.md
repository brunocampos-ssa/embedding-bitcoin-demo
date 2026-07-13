# Schema initialization

The workshop keeps its small idempotent schema beside the database bootstrap in `internal/platform/database/database.go`. Introduce numbered migrations before changing an already-deployed schema.
