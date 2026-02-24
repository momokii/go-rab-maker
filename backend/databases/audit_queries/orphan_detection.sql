-- ============================================================================
-- ORPHANED DATA DETECTION QUERIES
-- ============================================================================
-- These queries help detect data integrity issues where foreign keys reference
-- non-existent parent records. Run these queries periodically to audit data.
-- ============================================================================

-- 1. Check for project_work_items with invalid project_id
-- (Should return 0 rows if data is clean)
SELECT 'Orphaned work items (invalid project_id)' as check_type,
       COUNT(*) as orphan_count
FROM project_work_items pwi
WHERE NOT EXISTS (
    SELECT 1 FROM projects p WHERE p.project_id = pwi.project_id
);

-- 2. Check for project_work_items with invalid category_id
SELECT 'Orphaned work items (invalid category_id)' as check_type,
       COUNT(*) as orphan_count
FROM project_work_items pwi
WHERE NOT EXISTS (
    SELECT 1 FROM master_work_categories mwc WHERE mwc.category_id = pwi.category_id
);

-- 3. Check for project_work_items with invalid ahsp_template_id (when not NULL)
SELECT 'Orphaned work items (invalid ahsp_template_id)' as check_type,
       COUNT(*) as orphan_count
FROM project_work_items pwi
WHERE pwi.ahsp_template_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM ahsp_templates at WHERE at.template_id = pwi.ahsp_template_id
);

-- 4. Check for project_item_costs with invalid work_item_id
SELECT 'Orphaned item costs (invalid work_item_id)' as check_type,
       COUNT(*) as orphan_count
FROM project_item_costs pic
WHERE NOT EXISTS (
    SELECT 1 FROM project_work_items pwi WHERE pwi.work_item_id = pic.work_item_id
);

-- 5. Check for ahsp_material_components with invalid template_id
SELECT 'Orphaned material components (invalid template_id)' as check_type,
       COUNT(*) as orphan_count
FROM ahsp_material_components amc
WHERE NOT EXISTS (
    SELECT 1 FROM ahsp_templates at WHERE at.template_id = amc.template_id
);

-- 6. Check for ahsp_material_components with invalid material_id
SELECT 'Orphaned material components (invalid material_id)' as check_type,
       COUNT(*) as orphan_count
FROM ahsp_material_components amc
WHERE NOT EXISTS (
    SELECT 1 FROM master_materials mm WHERE mm.material_id = amc.material_id
);

-- 7. Check for ahsp_labor_components with invalid template_id
SELECT 'Orphaned labor components (invalid template_id)' as check_type,
       COUNT(*) as orphan_count
FROM ahsp_labor_components alc
WHERE NOT EXISTS (
    SELECT 1 FROM ahsp_templates at WHERE at.template_id = alc.template_id
);

-- 8. Check for ahsp_labor_components with invalid labor_type_id
SELECT 'Orphaned labor components (invalid labor_type_id)' as check_type,
       COUNT(*) as orphan_count
FROM ahsp_labor_components alc
WHERE NOT EXISTS (
    SELECT 1 FROM master_labor_types mlt WHERE mlt.labor_type_id = alc.labor_type_id
);

-- 9. Check for projects with invalid user_id
SELECT 'Orphaned projects (invalid user_id)' as check_type,
       COUNT(*) as orphan_count
FROM projects p
WHERE NOT EXISTS (
    SELECT 1 FROM users u WHERE u.user_id = p.user_id
);

-- 10. Check for master_materials with invalid user_id (when not NULL)
SELECT 'Orphaned materials (invalid user_id)' as check_type,
       COUNT(*) as orphan_count
FROM master_materials mm
WHERE mm.user_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM users u WHERE u.user_id = mm.user_id
);

-- 11. Check for master_labor_types with invalid user_id (when not NULL)
SELECT 'Orphaned labor types (invalid user_id)' as check_type,
       COUNT(*) as orphan_count
FROM master_labor_types mlt
WHERE mlt.user_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM users u WHERE u.user_id = mlt.user_id
);

-- 12. Check for master_work_categories with invalid user_id (when not NULL)
SELECT 'Orphaned work categories (invalid user_id)' as check_type,
       COUNT(*) as orphan_count
FROM master_work_categories mwc
WHERE mwc.user_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM users u WHERE u.user_id = mwc.user_id
);

-- 13. Check for ahsp_templates with invalid user_id (when not NULL)
SELECT 'Orphaned AHSP templates (invalid user_id)' as check_type,
       COUNT(*) as orphan_count
FROM ahsp_templates at
WHERE at.user_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM users u WHERE u.user_id = at.user_id
);

-- ============================================================================
-- DETAILED ORPHAN RECORD LISTS (uncomment to see specific orphaned records)
-- ============================================================================

-- List specific orphaned work items with invalid project_id
-- SELECT pwi.work_item_id, pwi.project_id, pwi.description
-- FROM project_work_items pwi
-- WHERE NOT EXISTS (SELECT 1 FROM projects p WHERE p.project_id = pwi.project_id);

-- List specific orphaned work items with invalid category_id
-- SELECT pwi.work_item_id, pwi.category_id, pwi.description
-- FROM project_work_items pwi
-- WHERE NOT EXISTS (SELECT 1 FROM master_work_categories mwc WHERE mwc.category_id = pwi.category_id);

-- List specific orphaned item costs with invalid work_item_id
-- SELECT pic.cost_id, pic.work_item_id, pic.item_name
-- FROM project_item_costs pic
-- WHERE NOT EXISTS (SELECT 1 FROM project_work_items pwi WHERE pwi.work_item_id = pic.work_item_id);
