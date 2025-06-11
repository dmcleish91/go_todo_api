# Go Todo API

A simple Todo API built with Go.

---

## 🚀 Quick Start

### 1. Build the Application

```sh
go build -o todoapi.exe ./cmd/
```

### 2. Run the Application

```sh
./todoapi.exe
```

---

## 🛠️ Development Mode

To run the development environment (without building):

```sh
go run ./cmd/
```

---

## 🔐 Setting Up Secrets

1. Create a `.env` file in the root directory of the project.
2. Add the required environment variables to the `.env` file. For example:

   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=yourusername
   DB_PASSWORD=yourpassword
   DB_NAME=tododb
   ```

---

## 🏃 Running the Application

1. Ensure the `.env` file is properly configured with the necessary secrets.
2. Build the application:

   ```sh
   go build -o todoapi.exe ./cmd/
   ```

3. Run the application:

   ```sh
   ./todoapi.exe
   ```
