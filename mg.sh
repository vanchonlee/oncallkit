#!/bin/bash

# Database Migration Script for SLAR API
# Usage: ./mg.sh [command] [options]

# Database configuration
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"slar"}
DB_NAME=${DB_NAME:-"slar"}
DB_PASSWORD=${DB_PASSWORD:-"slar"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Migration directory
MIGRATIONS_DIR="migrations"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if psql is available
check_psql() {
    if ! command -v psql &> /dev/null; then
        log_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi
}

# Check database connection
check_connection() {
    log_info "Checking database connection..."
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" &> /dev/null; then
        log_success "Database connection successful"
        return 0
    else
        log_error "Cannot connect to database. Please check your connection settings."
        log_info "Host: $DB_HOST, Port: $DB_PORT, User: $DB_USER, Database: $DB_NAME"
        return 1
    fi
}

# Create migrations table if not exists
create_migrations_table() {
    log_info "Creating migrations table if not exists..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    " &> /dev/null
}

# Get applied migrations
get_applied_migrations() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT version FROM schema_migrations ORDER BY version;
    " 2>/dev/null | sed 's/^ *//' | grep -v '^$'
}

# Get available migrations
get_available_migrations() {
    if [ -d "$MIGRATIONS_DIR" ]; then
        ls $MIGRATIONS_DIR/*.sql 2>/dev/null | sort | sed 's/.*\///' | sed 's/\.sql$//'
    fi
}

# Run a single migration
run_migration() {
    local migration_file="$1"
    local migration_name=$(basename "$migration_file" .sql)
    
    log_info "Running migration: $migration_name"
    
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration_file"; then
        # Record migration as applied
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
            INSERT INTO schema_migrations (version) VALUES ('$migration_name')
            ON CONFLICT (version) DO NOTHING;
        " &> /dev/null
        log_success "Migration $migration_name completed successfully"
        return 0
    else
        log_error "Migration $migration_name failed"
        return 1
    fi
}

# Show help
show_help() {
    echo "Database Migration Tool for SLAR API"
    echo ""
    echo "Usage: ./mg.sh [command] [options]"
    echo ""
    echo "Commands:"
    echo "  up                    Run all pending migrations"
    echo "  up [migration]        Run specific migration"
    echo "  status                Show migration status"
    echo "  create [name]         Create new migration file"
    echo "  reset                 Drop all tables and run all migrations"
    echo "  check                 Check database connection"
    echo "  help                  Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  DB_HOST              Database host (default: localhost)"
    echo "  DB_PORT              Database port (default: 5432)"
    echo "  DB_USER              Database user (default: slar)"
    echo "  DB_NAME              Database name (default: slar)"
    echo "  DB_PASSWORD          Database password (default: slar)"
    echo ""
    echo "Examples:"
    echo "  ./mg.sh up                           # Run all pending migrations"
    echo "  ./mg.sh up 001_create_alerts         # Run specific migration"
    echo "  ./mg.sh status                       # Show migration status"
    echo "  ./mg.sh create add_user_roles        # Create new migration"
    echo "  ./mg.sh reset                        # Reset database"
}

# Show migration status
show_status() {
    log_info "Migration Status"
    echo "=================="
    
    local applied_migrations=($(get_applied_migrations))
    local available_migrations=($(get_available_migrations))
    
    if [ ${#available_migrations[@]} -eq 0 ]; then
        log_warning "No migration files found in $MIGRATIONS_DIR/"
        return
    fi
    
    echo ""
    echo "Available Migrations:"
    for migration in "${available_migrations[@]}"; do
        if [[ " ${applied_migrations[@]} " =~ " ${migration} " ]]; then
            echo -e "  ${GREEN}✓${NC} $migration (applied)"
        else
            echo -e "  ${YELLOW}○${NC} $migration (pending)"
        fi
    done
    
    echo ""
    echo "Applied: ${#applied_migrations[@]} / ${#available_migrations[@]}"
}

# Run all pending migrations
run_all_migrations() {
    log_info "Running all pending migrations..."
    
    local applied_migrations=($(get_applied_migrations))
    local available_migrations=($(get_available_migrations))
    local pending_count=0
    
    for migration in "${available_migrations[@]}"; do
        if [[ ! " ${applied_migrations[@]} " =~ " ${migration} " ]]; then
            if run_migration "$MIGRATIONS_DIR/$migration.sql"; then
                ((pending_count++))
            else
                log_error "Migration failed. Stopping."
                exit 1
            fi
        fi
    done
    
    if [ $pending_count -eq 0 ]; then
        log_info "No pending migrations found"
    else
        log_success "Applied $pending_count migration(s)"
    fi
}

# Create new migration file
create_migration() {
    local migration_name="$1"
    if [ -z "$migration_name" ]; then
        log_error "Migration name is required"
        echo "Usage: ./mg.sh create [migration_name]"
        exit 1
    fi
    
    # Create migrations directory if not exists
    mkdir -p "$MIGRATIONS_DIR"
    
    # Generate timestamp
    local timestamp=$(date +"%Y%m%d%H%M%S")
    local filename="${timestamp}_${migration_name}.sql"
    local filepath="$MIGRATIONS_DIR/$filename"
    
    # Create migration file template
    cat > "$filepath" << EOF
-- Migration: $migration_name
-- Created: $(date)

-- Add your migration SQL here
-- Example:
-- CREATE TABLE example (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- Don't forget to add rollback instructions in comments:
-- ROLLBACK: DROP TABLE example;
EOF
    
    log_success "Created migration file: $filepath"
    log_info "Edit the file and add your SQL statements"
}

# Reset database
reset_database() {
    log_warning "This will drop all tables and re-run all migrations!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Dropping all tables..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
            DROP SCHEMA public CASCADE;
            CREATE SCHEMA public;
            GRANT ALL ON SCHEMA public TO $DB_USER;
            GRANT ALL ON SCHEMA public TO public;
        "
        log_success "Database reset completed"
        log_info "Running all migrations..."
        create_migrations_table
        run_all_migrations
    else
        log_info "Reset cancelled"
    fi
}

# Main script logic
main() {
    check_psql
    
    case "${1:-help}" in
        "up")
            if ! check_connection; then exit 1; fi
            create_migrations_table
            if [ -n "$2" ]; then
                # Run specific migration
                if [ -f "$MIGRATIONS_DIR/$2.sql" ]; then
                    run_migration "$MIGRATIONS_DIR/$2.sql"
                else
                    log_error "Migration file not found: $MIGRATIONS_DIR/$2.sql"
                    exit 1
                fi
            else
                # Run all pending migrations
                run_all_migrations
            fi
            ;;
        "status")
            if ! check_connection; then exit 1; fi
            create_migrations_table
            show_status
            ;;
        "create")
            create_migration "$2"
            ;;
        "reset")
            if ! check_connection; then exit 1; fi
            reset_database
            ;;
        "check")
            check_connection
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 