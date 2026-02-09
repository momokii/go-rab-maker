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
- Real-time cost calculations
- User authentication and session management
- Mobile-responsive design with DaisyUI and Tailwind CSS

## Tech Stack

- **Backend**: Go 1.21+ with Fiber web framework
- **Frontend**: Templ (Go templating) with HTMX
- **Database**: SQLite with modernc.org driver
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
├── main.go                 # Application entry point
├── go.mod
├── go.sum
└── README.md
```

## Setup Instructions

### Prerequisites
- Go 1.21 or higher
- SQLite3 (included with modernc.org/sqlite driver)
- Git

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd go-rab-maker
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Configure environment variables (optional):**
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

4. **Run the application:**
   ```bash
   # Run directly
   go run main.go

   # Or build and run
   go build -o rab-maker
   ./rab-maker
   ```

5. **Access the application:**
   - Open your browser to `http://localhost:3002`
   - Default login: username: `admin`, password: `admin123`
   - **Important**: Change the default admin password in production!

### Development Mode

For development with hot-reload:

```bash
# Install air (Go hot reload tool)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
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
# Change the port in .env file or set it directly
PORT=3003 go run main.go
```

**Database file not found:**
- The database will be created automatically on first run
- Check file permissions in the `databases/` directory

**Session/Authentication issues:**
- Clear browser cookies
- Check that `SECRET_KEY_JWT` is set in production
- Verify `ENV` is set correctly (development vs production)

**Cost calculations not working:**
- Verify AHSP templates are properly configured
- Check that materials and labor types have valid prices
- Review server logs for calculation errors

### Debug Mode

Enable debug logging:
```bash
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

4. **Database backups:**
   - Regularly backup the `databases/database.sqlite` file
   - Consider using a separate database server for production

5. **File permissions:**
   - Restrict write access to the database file
   - Use proper file permissions for the application directory

## Production Deployment

### Using Docker (Recommended)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o rab-maker

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rab-maker .
COPY --from=builder /app/databases ./databases

ENV PORT=3002
ENV ENV=production

EXPOSE 3002
CMD ["./rab-maker"]
```

### Using systemd

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

## Changelog

### Recent Improvements
- Fixed critical bugs in delete operations and search functionality
- Added input validation using validator tags
- Implemented proper error handling for cost calculations
- Fixed NULL handling for user_id in multi-tenant scenarios
- Added input sanitization to prevent issues with whitespace
- Fixed CookieSecure for localhost development
- Removed hardcoded delays for better performance
- Fixed directory naming inconsistencies
- Added structured logging utility
- Updated documentation

### Version
Current version: v1.1.0

---

**Built with Go + Fiber + Templ + SQLite**
