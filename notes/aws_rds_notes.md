### 1 Connect to RDS (SSL is REQUIRED on AWS RDS)

Use libpq connection string format (most reliable on Windows):

    psql "host=<RDS_ENDPOINT> port=5432 dbname=<DB_NAME> user=<DB_USER> sslmode=require"

Example:
    psql "host=jobping-db.ccf6io0y8fg6.us-east-1.rds.amazonaws.com port=5432 dbname=jobscanner user=jobscanner sslmode=require"

Enter the DB password when prompted.

---

### 2 Common psql commands

Inside psql:
    \\dt                 -- list tables
    \\d table_name       -- describe table
    SELECT * FROM table_name;
    \\q                  -- quit

    -- Verify connection context
    SELECT current_database();
    SELECT current_user;

    -- List schemas
    \\dn                 -- list schemas

    -- List tables
    \\dt                 -- list tables
    \\dt public.*        -- list tables in public schema

    -- Inspect table structure
    \\d table_name       -- describe table (columns, types, indexes)

    -- View data
    SELECT * FROM table_name;
    SELECT * FROM table_name LIMIT 20;
    SELECT * FROM table_name ORDER BY created_at DESC LIMIT 20;

    -- Count rows
    SELECT COUNT(*) FROM table_name;

    -- Filter data
    SELECT * FROM users WHERE email = 'test@example.com';

    -- Show columns only
    SELECT column_name, data_type
    FROM information_schema.columns
    WHERE table_name = 'table_name';

    -- List indexes
    \\di                 -- list indexes


