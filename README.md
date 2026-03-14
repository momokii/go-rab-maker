# RAB Maker

A comprehensive **Rencana Anggaran Biaya** (Budget Planning) application for construction projects. This application streamlines the cost estimation process for construction projects, making budget calculations easier, more structured, and clearly visible for all stakeholders.

## Features

### Core Functionality
- **Project Management**: Create and manage construction projects with detailed specifications
- **Work Items**: Add and manage work items with automatic cost calculations
- **Material Management**: Master data for construction materials with pricing
- **Labor Types**: Define labor types with daily wage rates
- **Work Categories**: Organize work items by category (e.g., foundation, structure, finishing)
- **AHSP Templates**: Create reusable cost templates based on standard unit prices
- **Cost Calculations**: Automatic material and labor cost calculations based on templates
- **Material Summaries**: Aggregate material requirements across projects with export functionality
- **Multi-User Support**: User-specific data with system-wide defaults

### Technical Highlights
- Server-side rendering with HTMX for responsive UX
- Real-time cost calculations with formatted currency display (Indonesian Rupiah)
- User authentication and session management
- Mobile-responsive design with DaisyUI and Tailwind CSS
- PDF and Excel export functionality for reports

## Tech Stack

- **Backend**: Go 1.24+ with Fiber web framework
- **Frontend**: Templ (Go templating) with HTMX
- **Database**: SQLite with modernc.org driver (pure Go, no CGo)
- **UI**: DaisyUI + Tailwind CSS
- **Authentication**: Session-based authentication

## Project Structure

```
go-rab-maker/
├── backend/
│   ├── databases/           # Database configuration and migrations
│   │   ├── migrations/     # SQL migration files
│   │   └── sqlite.go       # SQLite setup
│   ├── handlers/           # HTTP request handlers
│   ├── middlewares/        # Authentication and app middleware
│   ├── models/             # Data models and structures
│   ├── repository/         # Data access layer
│   │   ├── master_materials/
│   │   ├── master_labor_types/
│   │   ├── master_work_categories/
│   │   ├── ahsp_templates/
│   │   ├── projects/
│   │   ├── project_work_items/
│   │   └── ...
│   └── utils/              # Utility functions
├── frontend/
│   └── components/         # Templ components
│       ├── utils.go        # Shared utility functions
│       └── *.templ         # Template files
├── Dockerfile              # Docker build configuration
├── docker-compose.yml      # Docker Compose setup
├── .env.example           # Environment variables template
├── main.go                # Application entry point
├── go.mod
├── go.sum
└── README.md
```

## Quick Start

### Using Docker (Recommended)

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd go-rab-maker
   ```

2. **Configure environment variables (optional):**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run with Docker Compose:**
   ```bash
   docker-compose up -d
   ```

4. **Access the application:**
   - Open your browser to `http://localhost:3002`
   - Default login: username: `admin`, password: `admin123`
   - **Important**: Change the default admin password in production!

5. **Stop the application:**
   ```bash
   docker-compose down
   ```

### Manual Installation

#### Prerequisites
- Go 1.24 or higher
- Git
- Templ CLI (for template compilation)

#### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd go-rab-maker
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Install Templ CLI:**
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

4. **Build frontend templates:**
   ```bash
   templ generate
   ```

5. **Configure environment variables (optional):**
   ```bash
   # Create a .env file (optional - default values will be used)
   cat > .env << EOF
   PORT=3002
   ENV=development
   SECRET_KEY_JWT=your-secret-key-here
   EOF
   ```

   **Environment Variables:**
   - `PORT`: Server port (default: 3002)
   - `ENV`: Environment mode (`development` or `production`)
   - `SECRET_KEY_JWT`: Secret key for session encryption
   - `DEBUG`: Enable debug logging (`1` or `true`)

6. **Run the application:**
   ```bash
   # Run directly
   go run main.go

   # Or build and run
   go build -o rab-maker
   ./rab-maker
   ```

7. **Access the application:**
   - Open your browser to `http://localhost:3002`

## Development Mode

For development with hot-reload:

```bash
# Install air (Go hot reload tool)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

**Important:** After making changes to `.templ` files, you need to run:
```bash
templ generate
```

## Usage Guide

### Basic Workflow

1. **Setup Master Data:**
   - Navigate to Master Data in the sidebar
   - Add Materials (e.g., Cement, Sand, Steel)
   - Add Labor Types (e.g., Carpenter, Mason, Laborer)
   - Add Work Categories (e.g., Foundation, Structure, Finishing)

2. **Create AHSP Templates:**
   - Go to AHSP Templates
   - Create a new template (e.g., "1m³ Concrete Wall")
   - Add Material Components (e.g., Cement: 350 kg, Sand: 0.5 m³)
   - Add Labor Components (e.g., Mason: 8 hours, Laborer: 4 hours)

3. **Create a Project:**
   - Go to Projects
   - Create a new project with name, location, and client

4. **Add Work Items:**
   - Open a project detail page
   - Add work items and select AHSP templates
   - System automatically calculates costs based on template
   - Or enter manual costs for custom items

5. **View Material Summaries:**
   - Check the Material Summary page for aggregate requirements
   - Export to PDF or Excel for procurement planning

## Currency Formatting

The application displays all monetary values in Indonesian Rupiah format with thousand separators:
- **Format**: `Rp 1.000.000` (1 million Rupiah)
- **No decimals**: All amounts are rounded to whole numbers
- Applied across: Dashboard, Project Details, Material Summaries, and all cost displays

## Database Migrations

The application uses a custom migration system with embedded SQL files:

### Migration Files
Located in `backend/databases/migrations/`:
- `000001_basic_schema.up.sql` - Initial database schema
- `000001_basic_schema.down.sql` - Rollback for initial schema
- `000002_add_user_soft_delete.up.sql` - Adds soft delete capability
- `000002_add_user_soft_delete.down.sql` - Rollback soft delete
- `000003_add_unit_to_project_item_costs.up.sql` - Adds unit column
- `000003_add_unit_to_project_item_costs.down.sql` - Rollback unit column

### How Migrations Work
1. **Automatic on startup**: Migrations run automatically when the application starts
2. **Only on fresh database**: Migrations only run if the database file doesn't exist
3. **Embedded files**: Migration files are embedded in the binary using `go:embed`
4. **Up migrations only**: Only `.up.sql` files are executed during initialization
5. **Down migrations**: The `.down.sql` files are kept for reference and potential future manual rollback functionality

### Adding New Migrations
1. Create new migration files following the naming pattern: `000004_description.up.sql` and `000004_description.down.sql`
2. Place them in `backend/databases/migrations/`
3. Rebuild the application
4. To test, delete the database file and restart the application

**Important**: `.down.sql` files are NOT executed during initialization. They are kept for documentation and potential future rollback use.

## Database Schema

The application uses SQLite with the following main tables:

- `users` - User accounts and authentication
- `projects` - Construction projects
- `master_materials` - Material catalog with pricing
- `master_labor_types` - Labor types with daily wages
- `master_work_categories` - Work category definitions
- `ahsp_templates` - Reusable cost templates
- `ahsp_material_components` - Material components in templates
- `ahsp_labor_components` - Labor components in templates
- `project_work_items` - Work items within projects
- `project_item_costs` - Calculated costs for work items

## Development Guidelines

### Code Style
- Follow Go standard conventions
- Use meaningful variable and function names
- Add comments for complex business logic
- Keep functions focused and modular

### Validation
- All form inputs are validated using struct tags
- Validation errors are displayed to users in modals
- Input sanitization (trim whitespace) is applied to all form data

### Error Handling
- Errors are logged with context
- User-facing error messages are clear and actionable
- Database transactions are used for data integrity

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./backend/repository/master_materials/
```

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Change the port in .env file or docker-compose.yml
PORT=3003 docker-compose up -d
```

**Database file not found:**
- The database will be created automatically on first run
- Check file permissions in the `databases/` directory

**Volume permission errors (Docker):**
- The container includes an entrypoint script that automatically fixes permissions
- If you see "unable to open database file: out of memory (14)", it's a permissions issue
- The entrypoint script runs as root, fixes the database directory permissions, then switches to appuser
- Make sure the `su-exec` package is available in the Docker image (included in Dockerfile)

**Session/Authentication issues:**
- Clear browser cookies
- Check that `SECRET_KEY_JWT` is set in production
- Verify `ENV` is set correctly (development vs production)

**Cost calculations not working:**
- Verify AHSP templates are properly configured
- Check that materials and labor types have valid prices
- Review server logs for calculation errors

**Docker build issues:**
```bash
# Rebuild without cache
docker-compose build --no-cache

# Check logs
docker-compose logs -f
```

### Debug Mode

Enable debug logging:
```bash
# Docker
DEBUG=1 docker-compose up

# Manual
DEBUG=1 go run main.go
```

## Security Considerations

### For Production Deployment:

1. **Change default credentials:**
   - The default admin password is `admin123` - change this immediately!
   - Update the migration file or create a new admin user through the UI

2. **Use HTTPS:**
   - Set `ENV=production` in your environment
   - Configure a reverse proxy (nginx, Caddy) for SSL/TLS

3. **Set strong secret key:**
   - Generate a random `SECRET_KEY_JWT` and set it in environment variables
   - Use: `openssl rand -base64 32`

4. **Database backups:**
   - Regularly backup the database file from Docker volume
   - Use: `docker exec rab-maker-app cp /app/backend/databases/database.sqlite /backup/`
   - Or backup the named volume directly

5. **File permissions:**
   - The Docker container runs as non-root user `appuser`
   - Ensure proper permissions for mounted volumes

## Production Deployment

### Using Docker (Recommended)

The application is containerized and ready for production deployment with Docker Compose.

**Key Features:**
- Multi-stage build with Alpine Linux for minimal image size
- Non-root user (`appuser`) for security
- Automatic permission handling via entrypoint script
- Health checks for container monitoring
- Persistent volume for database storage

**Deployment Steps:**

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

**Database Persistence:**
- The database is stored in a Docker volume at `/app/databases/database.sqlite`
- A bind mount `./rab-db-data:/app/databases` is used for easy local access
- The entrypoint script automatically fixes permissions on startup
- Data persists across container restarts

**Note on the "OOM" Error:**
If you see "unable to open database file: out of memory (14)", this is NOT an actual memory issue. It's SQLite error code 14 (SQLITE_CANTOPEN), which occurs when:
- The database directory has incorrect permissions
- The container runs as `appuser` (uid 1000) but the directory is owned by root

The entrypoint script (`docker-entrypoint.sh`) handles this automatically by:
1. Running as root initially
2. Creating/fixing the database directory permissions
3. Switching to `appuser` before starting the application

### Using systemd

For traditional server deployment:

Create a service file at `/etc/systemd/system/rab-maker.service`:

```ini
[Unit]
Description=RAB Maker Application
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/rab-maker
ExecStart=/opt/rab-maker/rab-maker
Restart=always
Environment="PORT=3002"
Environment="ENV=production"

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl enable rab-maker
sudo systemctl start rab-maker
```

## Changelog

### v1.3.0 (Latest)

**Docker Deployment Fixes:**
- Fixed Docker volume permission issue that caused "OOM" error
- Added `docker-entrypoint.sh` script for automatic permission handling
- Container now starts as root, fixes database permissions, then switches to appuser
- Added `su-exec` package for privilege dropping
- Database now persists correctly with bind mount `./rab-db-data:/app/databases`

**Note on "OOM" Error:**
- The "unable to open database file: out of memory (14)" error was NOT a real memory issue
- It was SQLite error code 14 (SQLITE_CANTOPEN) due to permission problems
- Fixed by ensuring the database directory is writable by appuser (uid 1000)
- The entrypoint script handles this automatically on container startup

### v1.2.0

**New Features:**
- Currency formatting with Indonesian thousand separators (Rp 1.000.000)
- Enhanced dashboard with improved statistics and breakdowns
- Added `frontend/components/utils.go` for shared utility functions

**Improvements:**
- Enhanced PDF/Excel export functionality
- Fixed popup error modals throughout the application
- Added "every page explanation" sections for better UX
- Removed debug log statements for cleaner production logs

**Bug Fixes:**
- Fixed modal errors when deleting master data that is in use
- Fixed manual entry display and editing for work items
- Fixed project detail views and summaries
- Fixed search functionality across tables

### v1.1.0
- Fixed critical bugs in delete operations and search functionality
- Added input validation using validator tags
- Implemented proper error handling for cost calculations
- Fixed NULL handling for user_id in multi-tenant scenarios
- Added input sanitization to prevent issues with whitespace
- Fixed CookieSecure for localhost development
- Removed hardcoded delays for better performance
- Fixed directory naming inconsistencies
- Added structured logging utility

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with clear commit messages
4. Add tests for new functionality
5. Submit a pull request

## License

[Specify your license here]

## Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section above
- Review the code comments and documentation

---

**Built with Go + Fiber + Templ + SQLite**
