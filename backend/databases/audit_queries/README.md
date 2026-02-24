# Database Audit Queries

This directory contains SQL queries for auditing and maintaining database integrity.

## Purpose

These queries are **diagnostic tools** for:
- Detecting orphaned data (foreign keys pointing to non-existent records)
- Validating data integrity
- Periodic health checks

**Note:** These are NOT migration files. They are read-only queries that don't modify the database.

## Usage

### Running Individual Queries

```bash
# Connect to your SQLite database
sqlite3 /path/to/database.sqlite

# Read and run specific query from the file
sqlite3 /path/to/database.sqlite < orphan_detection.sql
```

### Running with sqlite3 CLI

```bash
# Open database
sqlite3 databases/database.sqlite

# Run a specific check (example: check for orphaned work items)
SELECT 'Orphaned work items (invalid project_id)' as check_type,
       COUNT(*) as orphan_count
FROM project_work_items pwi
WHERE NOT EXISTS (
    SELECT 1 FROM projects p WHERE p.project_id = pwi.project_id
);
```

### Expected Results

All queries should return **orphan_count = 0** for a healthy database.

If any query returns a non-zero count, investigate the orphaned records.

## Files

| File | Purpose |
|------|---------|
| `orphan_detection.sql` | Checks for orphaned records across all foreign key relationships |

## Query Categories

### 1. Project Work Items
- Validates `project_id` references
- Validates `category_id` references
- Validates `ahsp_template_id` references (when set)

### 2. Project Item Costs
- Validates `work_item_id` references

### 3. AHSP Components
- Validates `template_id` references
- Validates `material_id` / `labor_type_id` references

### 4. Master Data
- Validates `user_id` references for all master tables

## When to Run

- **After bulk data operations** (imports, deletions)
- **Before deploying to production**
- **Periodic maintenance** (weekly/monthly)
- **After schema changes**

## Interpreting Results

```sql
-- Example output:
check_type                                    | orphan_count
----------------------------------------------|-------------
Orphaned work items (invalid project_id)      | 0
Orphaned work items (invalid category_id)     | 0
...
```

- **orphan_count = 0**: No issues found
- **orphan_count > 0**: Investigate the specific records

### Finding Specific Orphaned Records

Uncomment the detailed queries at the bottom of `orphan_detection.sql` to see specific orphaned records:

```sql
-- List specific orphaned work items
SELECT pwi.work_item_id, pwi.project_id, pwi.description
FROM project_work_items pwi
WHERE NOT EXISTS (SELECT 1 FROM projects p WHERE p.project_id = pwi.project_id);
```

## Best Practices

1. **Run in a transaction** if investigating and fixing
2. **Backup before cleanup** - Always backup before deleting orphaned records
3. **Understand the root cause** - Don't just delete; find why orphans exist
4. **Document findings** - Keep a record of data integrity issues found
