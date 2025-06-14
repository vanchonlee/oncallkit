psql -U slar -d slar -h localhost -f 001_create_alerts.sql
psql -U slar -d slar -h localhost -f migrations/002_add_alert_columns.sql