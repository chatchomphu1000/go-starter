# /new-migration — Create a MongoDB migration pair

Create a new golang-migrate migration pair for this project.
Argument: `$ARGUMENTS` — the migration description (e.g. `add_product_indexes`, `add_orders_collection`)

## Steps

### 1. Determine next sequence number
```bash
ls migrations/ | grep -E '^[0-9]+' | sort | tail -1
```
Increment the numeric prefix by 1. Zero-pad to 6 digits (e.g. `000002`).

### 2. Normalize the migration name
- Replace spaces with `_`
- Lowercase everything
- Final name: `{SEQ}_{description}` (e.g. `000002_add_product_indexes`)

### 3. Infer collection name from description
Parse `$ARGUMENTS` to extract the collection name:
- If it contains `create_X_indexes` or `add_X_indexes`, the collection is `X`
- If it contains `add_X_collection`, the collection is `X`
- Otherwise, ask the user to clarify which collection this migration targets

### 4. Generate the `.up.json` file
Path: `migrations/{SEQ}_{description}.up.json`

For index migrations, scaffold:
```json
{
  "createIndexes": "{collection}",
  "indexes": [
    {
      "key": { "field_name": 1 },
      "name": "field_name_idx",
      "background": false
    }
  ]
}
```

For collection creation:
```json
{
  "create": "{collection}",
  "validator": {
    "$jsonSchema": {
      "bsonType": "object",
      "required": ["_id", "created_at"],
      "properties": {
        "_id":        { "bsonType": "string", "description": "UUID v7 string ID" },
        "created_at": { "bsonType": "date" },
        "updated_at": { "bsonType": "date" }
      }
    }
  }
}
```

### 5. Generate the `.down.json` file
Path: `migrations/{SEQ}_{description}.down.json`

For index migrations:
```json
{
  "dropIndexes": "{collection}",
  "index": ["field_name_idx"]
}
```

For collection:
```json
{
  "drop": "{collection}"
}
```

### 6. Verify the golang-migrate format
Confirm that:
- The up file is valid JSON: `cat migrations/{file}.up.json | python3 -m json.tool`
- The down file is valid JSON: `cat migrations/{file}.down.json | python3 -m json.tool`

### 7. Report to user
Print:
```
Created:
  migrations/{SEQ}_{description}.up.json
  migrations/{SEQ}_{description}.down.json

Run: make migrate-up        → apply
     make migrate-status    → verify
     make migrate-down      → rollback (1 step)

Note: Edit the index/field definitions in the .up.json before running.
```

## Project migration conventions
- `_id` is always a **string** (UUID v7), never ObjectID
- Index names follow `{field}_{type}` pattern (e.g. `email_unique`, `created_at_desc`)
- Unique indexes use `"unique": true`
- Case-insensitive text indexes use `"collation": { "locale": "en", "strength": 2 }`
- Background index builds: `"background": false` (MongoDB 4.2+ always builds in background)
