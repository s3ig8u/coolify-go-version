# Database Migration Guide

## 🔄 **Migration System Overview**

Coolify Go uses **GORM Auto-Migration** for database schema management. This system automatically creates and updates database tables based on your Go struct definitions.

## 🚀 **How It Works**

### **1. Automatic Migration (Current System)**

When the application starts, GORM automatically:

-   ✅ Creates tables if they don't exist
-   ✅ Adds new columns when struct fields are added
-   ✅ Updates column types when struct types change
-   ✅ Creates indexes based on GORM tags
-   ✅ Establishes foreign key relationships

### **2. Migration Triggers**

Migrations run automatically in these scenarios:

#### **Application Startup**

```go
// In main.go
if err := config.ConnectDatabase(appConfig.Database); err != nil {
    log.Printf("Database connection failed: %v", err)
} else {
    // Run database migrations
    if err := config.AutoMigrate(); err != nil {
        log.Printf("Database migration failed: %v", err)
    }
}
```

#### **Manual Migration Command**

```bash
# Run migrations manually
make migrate

# Check migration status
make migrate-status

# Rollback migrations (placeholder)
make migrate-rollback
```

## 📋 **Migration Commands**

### **Available Commands:**

```bash
# Run migrations
make migrate
# or
go run cmd/migrate/main.go -action=migrate

# Check migration status
make migrate-status
# or
go run cmd/migrate/main.go -action=status

# Rollback (placeholder)
make migrate-rollback
# or
go run cmd/migrate/main.go -action=rollback -step=1
```

### **Docker Environment:**

```bash
# Run migrations in Docker container
docker exec coolify-go ./coolify-go -migrate

# Check status in Docker container
docker exec coolify-go ./coolify-go -migrate-status
```

## 🏗️ **Model-Based Migrations**

### **How Models Define Schema:**

```go
type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Email     string         `json:"email" gorm:"uniqueIndex;not null"`
    Password  string         `json:"-" gorm:"not null"`
    Name      string         `json:"name" gorm:"not null"`
    Role      string         `json:"role" gorm:"default:'user'"`
    IsActive  bool           `json:"is_active" gorm:"default:true"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
```

**GORM Tags Explained:**

-   `primaryKey` - Sets as primary key
-   `uniqueIndex` - Creates unique index
-   `not null` - Makes column NOT NULL
-   `default:'value'` - Sets default value
-   `index` - Creates regular index
-   `foreignKey:UserID` - Establishes foreign key relationship

### **Generated SQL (Example):**

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(255) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE UNIQUE INDEX idx_users_email ON users(email);
```

## 🔧 **Adding New Models**

### **Step 1: Create Model File**

```go
// internal/models/new_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type NewModel struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Name      string         `json:"name" gorm:"not null"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### **Step 2: Add to Migration**

```go
// internal/config/database.go
func AutoMigrate() error {
    return DB.AutoMigrate(
        &models.User{},
        &models.Team{},
        &models.Server{},
        &models.Application{},
        &models.NewModel{}, // Add new model here
    )
}
```

### **Step 3: Run Migration**

```bash
make migrate
```

## ⚠️ **Important Considerations**

### **1. Data Safety**

-   **Auto-migration is safe** for adding new columns
-   **Column type changes** may cause data loss
-   **Column deletions** will remove data
-   **Always backup** before major schema changes

### **2. Production Deployment**

```bash
# Safe migration process
1. Backup database
2. Deploy new code
3. Run migrations
4. Verify application
5. Rollback if issues
```

### **3. Development Workflow**

```bash
# Local development
1. Modify model structs
2. Run: make migrate
3. Test application
4. Commit changes
```

## 🔄 **Migration Best Practices**

### **1. Model Changes**

-   ✅ **Safe**: Adding new fields with defaults
-   ✅ **Safe**: Adding new models
-   ✅ **Safe**: Adding indexes
-   ⚠️ **Careful**: Changing field types
-   ❌ **Dangerous**: Removing fields (data loss)

### **2. Relationship Changes**

```go
// Safe: Adding relationships
type User struct {
    // ... existing fields
    NewRelation []NewModel `json:"new_relation" gorm:"foreignKey:UserID"`
}

// Careful: Changing relationship types
// May require manual migration
```

### **3. Index Management**

```go
// Adding indexes is safe
type User struct {
    Email string `json:"email" gorm:"uniqueIndex;not null"`
    Name  string `json:"name" gorm:"index"` // New index
}
```

## 🚨 **Troubleshooting**

### **Common Issues:**

#### **1. Migration Fails**

```bash
# Check database connection
make migrate-status

# Check logs
docker logs coolify-go

# Manual connection test
docker exec coolify-go-db psql -U coolify_go -d coolify_go -c "\dt"
```

#### **2. Schema Mismatch**

```bash
# Reset database (development only)
docker-compose down
docker volume rm coolify_postgres_data
docker-compose up -d
make migrate
```

#### **3. Data Loss Prevention**

```bash
# Backup before migration
docker exec coolify-go-db pg_dump -U coolify_go coolify_go > backup.sql

# Restore if needed
docker exec -i coolify-go-db psql -U coolify_go coolify_go < backup.sql
```

## 🔮 **Future Improvements**

### **Planned Migration System:**

1. **Versioned Migrations** - Track migration versions
2. **Rollback Support** - Proper rollback functionality
3. **Migration Files** - SQL-based migration files
4. **Migration History** - Track applied migrations
5. **Dry Run Mode** - Preview migration changes

### **Example Future Migration File:**

```sql
-- migrations/001_create_users_table.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- migrations/002_add_user_role.sql
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user';
```

---

**Current System:** GORM Auto-Migration  
**Status:** Production Ready  
**Next Enhancement:** Versioned Migration System
