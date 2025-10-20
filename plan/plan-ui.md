# 🎨 UI/UX Design & Workflow Plan

This document details the user interface design plan, navigation structure, and UI implementation roadmap for the Simple BoQ Application. Its purpose is to provide a visual and functional guide that aligns with the technical development plan.

---

## 🏛️ Navigation Structure & Sidebar Menu Recommendation

To keep the application simple and intuitive, the main navigation will be placed in a persistent sidebar. The recommended menu structure is as follows:

* **`Dashboard`**
    * The main landing page that displays a list of all ongoing projects. This serves as the starting point for the user.

* **`Master Data`** (Menu Group)
    * **`Materials`**: The page for managing all building material data.
    * **`Work Templates`**: The page for managing the Unit Price Analysis or work "recipes".

* **`About`**
    * A simple page containing information about the application version, creator, or help contacts. (Low priority).

---

## 🗺️ UI Implementation Roadmap

The development order for the interface is designed to ensure functionalities are built logically and incrementally, in line with the technical roadmap.

1.  **Step 1: The Foundation - Master Data Pages**
    * **Goal**: To build the application's "kitchen" first. Users must be able to input basic data (materials and work recipes) before they can create a BoQ.
    * **Pages to Create**:
        * UI for `Materials` CRUD.
        * UI for `Work Templates` CRUD.
        * UI to manage Material Requirements (Coefficients) within each Template.

2.  **Step 2: The Main Framework - Dashboard & Project Pages**
    * **Goal**: To create the "containers" for each BoQ job.
    * **Pages to Create**:
        * Dashboard UI to display a list of projects as cards or a table.
        * A form (in a modal/pop-up) to create and edit new Projects.

3.  **Step 3: The Core Application - Project Detail Page**
    * **Goal**: To implement the main feature where users will spend most of their time. This is the most complex page.
    * **Pages/Components to Create**:
        * The main layout for the Project Detail Page.
        * Functionality for "Add Work Item," including a form with an interactive volume calculator.
        * A list display of added work items, which can be expanded to show the calculated material details.
        * A form to edit material details (custom prices/coefficients).

4.  **Step 4: Reporting & Output - Summary Page**
    * **Goal**: To provide a useful and actionable output for the user.
    * **Pages to Create**:
        * The "Material Summary" tab/page view, containing a "shopping list" summary table.
        * Implementation of the "Export" button to trigger a file download.

---

## 🖥️ Design & Functionality Details per Page

Here are the details of the components, data, and user actions for each page to be created.

### **1. Page: Master Materials (`/master/materials`)**

* **UI Components**:
    * Page Title: "Master Materials"
    * Main Action Button: **`+ Add New Material`**
    * Search Bar: To find materials by name.
    * Materials Table.

* **Data Displayed (in Table)**:
    * `Material Name` (from `materials.name`)
    * `Unit` (from `materials.unit`)
    * `Default Price` (from `materials.default_price`), formatted as currency.
    * `Actions` column.

* **User Actions**:
    * **Click `+ Add New Material`**: Opens a modal/form with inputs for Name, Unit, and Default Price. Saving triggers an `INSERT` into the `materials` table.
    * **Click `Edit` (pencil icon) on a table row**: Opens the same modal, but pre-filled with existing data. Saving triggers an `UPDATE`.
    * **Click `Delete` (trash icon) on a table row**: Displays a confirmation dialog. If confirmed, triggers a `DELETE` from the `materials` table.

### **2. Page: Master Work Templates (`/master/work-templates`)**

* **UI Components**:
    * Page Title: "Master Work Templates"
    * Main Action Button: **`+ Add New Template`**
    * Work Templates Table.

* **Data Displayed (in Table)**:
    * `Work Name` (from `work_types.name`)
    * `Category` (from `work_types.category`)
    * `Unit` (from `work_types.unit`)
    * `Actions` column.

* **User Actions**:
    * **Basic CRUD**: Similar to the Master Materials page for the `work_types` table.
    * **Click on a Work Name (or a "details" icon)**: This is the main action. The user will be navigated to the detail page to manage its recipe (`/master/work-templates/{id}`).

### **3. Page: Work Template Details (`/master/work-templates/{id}`)**

* **UI Components**:
    * Header: Displays the `Work Name`, `Category`, and `Unit`.
    * Sub-Header: "Material Requirements (per 1 {unit})"
    * Action Button: **`+ Add Material Requirement`**
    * Material Requirements Table.

* **Data Displayed (in Table)**:
    * `Material Name` (from `materials.name`)
    * `Coefficient` (from `coefficients.value`)
    * `Unit` (from `materials.unit`)
    * `Actions` column.

* **User Actions**:
    * **Click `+ Add Material Requirement`**: Opens a modal with:
        * A dropdown to select a material (from `SELECT * FROM materials`).
        * A number input for the `Coefficient Value`.
        * Saving will `INSERT` into the `coefficients` table.
    * **Edit/Delete** on each row to manage the existing recipe.

### **4. Page: Dashboard (`/`)**

* **UI Components**:
    * Page Title: "My Projects"
    * Main Action Button: **`+ Create New Project`**
    * Content Area: Displays a list of projects as cards or table rows.

* **Data Displayed (per Project)**:
    * `Project Name` (from `projects.name`)
    * `Location` (from `projects.location`)
    * `Date Created` (from `projects.created_at`)

* **User Actions**:
    * **Click `+ Create New Project`**: Opens a modal with inputs for Name, Location, Description. Saving will `INSERT` into `projects`.
    * **Click on a project card**: Navigates the user to the **Project Detail Page** (`/project/{id}`).

### **5. Page: Project Detail (`/project/{id}`)**

* **UI Components**:
    * *Project Header*: Displays the Project Name and Total Estimated Cost (calculated in real-time).
    * Tab Navigation: **`BoQ`** | **`Material Summary`**
    * Main Action Button: **`+ Add Work Item`**
    * Content Area (BoQ Tab): A list of work items, grouped by category (`Wall Works`, `Floor Works`, etc.). Each item should be expandable/collapsible.

* **Data Displayed (when item is expanded)**:
    * A small table containing the required material details.
    * Columns: `Material`, `Coefficient`, `Unit Price`, `Requirement`, `Subtotal Cost`.
    * This data is fetched from `project_material_details` and related tables.

* **User Actions**:
    * **Click `+ Add Work Item`**: Opens a modal with:
        * A dropdown to select a `Work Template` (from `work_types`).
        * Interactive inputs for Volume (e.g., Length x Width).
        * Saving will `INSERT` into `project_works` and trigger the calculation on the backend.
    * **Click `Edit` on a material detail**: Allows the user to change the `unit_price` or `coefficient` specifically for that item (triggers an `UPDATE` on `project_material_details`).
    * **Click the `Material Summary` tab**: Displays the content from the Summary Page.

### **6. Content Tab: Material Summary**

* **UI Components**:
    * Title: "Material Requirement Summary (Shopping List)"
    * Action Button: **`Export to PDF`**
    * Summary Table.

* **Data Displayed (in Table)**:
    * `Material Name`
    * `Total Requirement` (result of the `SUM` calculation)
    * `Unit`

* **User Actions**:
    * **Click `Export to PDF`**: Triggers the backend to generate the file and initiates the download process.