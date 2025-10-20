# 📖 Final Development Plan: Simple BoQ Application (v3 - Efficiency Focused)

This document is the final, most comprehensive version of the development and UI/UX plan, fully aligned with the `000001_basic_schema.up.sql` schema. This version has been revised with a strong focus on **development efficiency** by specifying exact database operations (CRUD) needed for each feature and detailing all **automated process flows**.

---

## ✨ Application Architecture & Features (Confirmed)

The architecture remains the same (Multi-User, Master Data, Core BoQ, Cost Snapshotting). This plan will detail *how* to implement it efficiently.

---

## ⚙️ Automated Process Flows & Business Rules

This new section explicitly details the "cause-and-effect" logic that the backend must handle. Understanding this is key to efficient development.

1.  **Creation of a Work Item (The Main Trigger)**
    * **WHEN** a user creates a new record in `project_work_items` and provides an `ahsp_template_id`.
    * **THEN** the system **MUST AUTOMATICALLY**:
        1.  **Read** all associated component "recipes" from `ahsp_material_components` and `ahsp_labor_components` using the `ahsp_template_id`.
        2.  **Loop** through each component found.
        3.  **Calculate** the required values (`quantity_needed`, `total_cost`).
        4.  **Create** multiple new records in the `project_item_costs` table, one for each material and labor component.

2.  **Updating a Work Item's Volume**
    * **WHEN** a user updates the `volume` on an existing `project_work_items` record.
    * **THEN** the system **MUST AUTOMATICALLY**:
        1.  **Delete** all existing records in `project_item_costs` that are linked to the updated `work_item_id`. This is crucial for preventing data duplication.
        2.  **Re-run** the entire creation process described in point #1 using the new `volume` to generate a fresh set of cost records.

3.  **Deleting a Work Item**
    * **WHEN** a user deletes a `project_work_items` record.
    * **THEN** the database's `ON DELETE CASCADE` constraint on the foreign key **WILL AUTOMATICALLY** delete all associated records in `project_item_costs`. You do not need to write separate backend logic for this if the schema is implemented correctly.

4.  **Changing a Master Price**
    * **WHEN** a user updates `default_unit_price` in `master_materials`.
    * **THEN** no other records are affected. The `project_item_costs` table intentionally holds the old price (`unit_price_at_creation`) to preserve the integrity of existing BoQs. This is a core business rule.

---

## 🛠️ Revised Development Roadmap (with Efficiency Notes)

### **Phase 0: Initial Setup**

1.  **Project & Database Setup**: Unchanged.
2.  **User Authentication**: Implement user login/registration. This is a prerequisite for everything else.

### **Phase 1: Building the Foundation (Master Data)**

1.  **Implement Full CRUD for all Master Tables**:
    * **Tables**: `master_materials`, `master_labor_types`, `master_work_categories`.
    * **Operations**: These tables require full **Create, Read, Update, Delete (CRUD)** functionality as the user needs complete control over their personal master data.

2.  **Implement Full CRUD for AHSP Templates & Components**:
    * **Tables**: `ahsp_templates`, `ahsp_material_components`, `ahsp_labor_components`.
    * **Operations**: These also require full **CRUD**. A user must be able to create a template, add components (materials/labor), update coefficients, and remove components.

### **Phase 2: Project Management**

1.  **Implement Full CRUD for `projects`**:
    * **Operations**: A user needs full **CRUD** to manage their project portfolio. This is straightforward.

### **Phase 3: Implementing Core BoQ Logic (Revised with Efficiency)**

1.  **Implement Full CRUD for `project_work_items`**:
    * **Operations**: The user must have full **CRUD** control over the work items within their project (adding a task, changing its description/volume, or deleting it).

2.  **Implement the Cost Snapshot Logic**:
    * **Table**: `project_item_costs`.
    * **Operations**: This is the key efficiency point. This table does **NOT** require a full user-facing CRUD module.
        * **Create**: **System-Only**. Records are created *only* by the automated trigger described in the "Automated Process Flows" section. There should be no UI button for "Add New Cost Item".
        * **Read**: **User-Facing**. The UI will read from this table to display the cost breakdown. This is its primary purpose from a user's perspective.
        * **Update**: **Not Required**. To maintain data integrity, individual cost items should not be editable. If a change is needed, it should happen on the parent `project_work_items` (e.g., changing volume), which will trigger a Delete-and-Recreate flow. This simplifies the logic immensely.
        * **Delete**: **System-Only**. Records are deleted *only* when the parent `project_work_items` is updated or deleted.
    * **Efficiency Note**: By making this table "Create/Delete by System, Read by User", you eliminate the need to build complex forms for editing individual cost components, saving significant development time.

### **Phase 4: Reporting & Display**

1.  **Implement Read-Only Displays**:
    * **Operations**: All reporting features (Project Detail View, Material Summary) are **Read-Only**. They execute `SELECT` queries (often with `JOIN`s, `SUM`s, and `GROUP BY`s) on the existing data and present it to the user. No `Create`, `Update`, or `Delete` operations are initiated from these views.

2.  **Implement Export Functionality**:
    * **Operations**: This is also a **Read-Only** operation. The system reads data and formats it into a file.

---

## 🎨 Revised UI/UX Design & Workflow Plan (with Efficiency Notes)

### **Sidebar Navigation Structure (Confirmed)**

* `Dashboard / Projects`
* `Master Data` (Group) -> `Materials`, `Labor Types`, `Work Categories`, `AHSP Templates`
* `Settings / Logout`

### **Design & Functionality Details per Page (Revised)**

* **Pages: All under `Master Data`**
    * **User Actions**: These pages will feature a full suite of user actions: "Add New", "Edit", and "Delete" buttons, along with a data table and search functionality.

* **Page: Project Detail (`/project/{id}`)**
    * **UI Components & Flow**:
        1.  The user clicks **"+ Add Work Item"**.
        2.  A form appears to create a `project_work_items` record.
        3.  Upon saving, the UI refreshes. The new work item appears in the list.
        4.  The user can expand this item to see its cost breakdown.
    * **Efficiency Note on UI**: The displayed cost breakdown (from `project_item_costs`) should be presented as a **read-only table**. There should be **NO "Edit" or "Delete" buttons on individual cost line items**. This prevents user confusion and simplifies the UI. If a user wants to change the costs, they should click "Edit" on the parent `project_work_items` container, which will re-trigger the automated calculation.

* **Editing a Work Item Flow**:
    1.  User clicks the "Edit" button on a `project_work_items` entry (e.g., "Pekerjaan dinding area depan").
    2.  The same form used for creation appears, pre-filled with the data.
    3.  User changes the `Volume` from 10 to 15 and clicks "Save".
    4.  **Backend Trigger**: The system executes the "Update" flow: deletes all old `project_item_costs` for this item and creates new ones based on the volume of 15.
    5.  **UI Refresh**: The UI re-fetches the data and now displays the updated cost breakdown, with all quantities and totals recalculated automatically.

* **Page/Tab: Material Summary**
    * **UI Components**: A simple, **read-only** table and an "Export" button.
    * **User Actions**: The only action here is to view the data or to click "Export". There is no data manipulation on this screen.