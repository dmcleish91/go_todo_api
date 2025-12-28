# Go Todo API

## Overview

A simple Todo API built with Go. This project provides a backend service for a todo application, allowing users to manage projects and tasks. It uses the Echo web framework and connects to a PostgreSQL database. **Currently running in demo mode** - authentication is disabled and all requests use a hardcoded demo user ID.

## Getting Started

### Prerequisites

- **Go**: Version 1.22 or newer ([Installation Guide](https://go.dev/doc/install))
- **PostgreSQL**: A running instance of PostgreSQL. ([Download](https://www.postgresql.org/download/))

### Setup

1.  **Clone the repository**
    ```bash
    git clone https://github.com/dmcleish91/go_todo_api.git
    cd go_todo_api
    ```

2.  **Setup environment variables**

    Create a `.env` file in the root directory of the project. You can copy the example below and fill in your details.

    ```env
    # .env file
    
    # PostgreSQL connection details
    host=localhost
    port=5432
    user=your_db_user
    password=your_db_password
    dbname=your_db_name
    
    # Demo mode: User ID for demo user (UUID from your database)
    DEMO_USER_ID=your-demo-user-uuid
    ```

3.  **Set up the database**

    Connect to your PostgreSQL instance and run the table creation queries from the `queries.sql` file to set up the necessary tables (`projects`, `tasks`, and `labels`).

    **Important for Demo Mode:** The foreign key constraints on `user_id` columns have been removed. If you're using existing tables that still have these constraints, you'll need to drop them:

    ```sql
    ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_user_id_fkey;
    ALTER TABLE labels DROP CONSTRAINT IF EXISTS labels_user_id_fkey;
    ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_user_id_fkey;
    ```

### Running the Development Server

To run the application in development mode:

```bash
go run ./cmd/
```

The application should now be available at: `http://localhost:1323`

### Running Tests

The project currently lacks a dedicated test suite. To run tests (once added):

```bash
go test ./...
```

## Building for Production

To build a production binary:

```bash
# For Linux/macOS
go build -o todoapi ./cmd/

# For Windows
go build -o todoapi.exe ./cmd/
```

The executable will be created in the root directory.

## Future Features / TODO

- [ ] Add a comprehensive test suite (unit and integration tests).
- [ ] Refactor environment variable handling to use a struct.
- [ ] Implement more sophisticated input validation.
- [ ] Add swagger documentation for the API endpoints.
- [ ] Implement soft-delete for tasks and projects.

## Demo Mode

The API is currently running in **demo mode**, which means:

- **No authentication required** - All endpoints are publicly accessible
- **Hardcoded user ID** - All operations use the `DEMO_USER_ID` from your environment variables
- **Reset endpoint** - Use `POST /v1/reset` to delete all demo user data and seed fresh sample data

### Reset Demo Data Endpoint

```
POST /v1/reset
```

This endpoint will:
1. Delete all existing tasks, labels, and projects for the demo user
2. Seed fresh demo data including:
   - An "Inbox" project
   - Three labels: "urgent", "work", "personal"
   - Five sample tasks with various priorities and due dates

**Example:**
```bash
curl -X POST http://localhost:1323/v1/reset
```

## Task Ordering

Tasks now have an `order` integer field, which determines their position among sibling tasks (same project and parent_task_id). You can reorder tasks using the new endpoint:

### Reorder Tasks Endpoint

```
PATCH /v1/tasks/reorder
```

**Request Body:**

```
[
  { "task_id": "a", "order": 0 },
  { "task_id": "b", "order": 1 },
  ...
]
```

- All tasks must belong to the demo user and have the same `project_id` and `parent_task_id`.
- The endpoint will update the order of these sibling tasks atomically.

## Contributing

1.  Fork the project.
2.  Create your feature branch (`git checkout -b feature/YourFeature`).
3.  Commit your changes (`git commit -m 'Add some feature'`).
4.  Push to the branch (`git push origin feature/YourFeature`).
5.  Open a Pull Request.
