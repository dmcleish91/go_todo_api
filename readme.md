# How to build and run the application

go build -o todoapi.exe .\cmd\ && .\todoapi.exe

run the dev env with go run .\cmd\

# Setting up secrets

1. Create a `.env` file in the root directory of the project.
2. Add the required environment variables to the `.env` file. For example:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=yourusername
   DB_PASSWORD=yourpassword
   DB_NAME=tododb
   ```

# Running the application

1. Ensure the `.env` file is properly configured with the necessary secrets.
2. Build the application:
   ```
   go build -o todoapi.exe .\cmd\
   ```
3. Run the application:
   ```
   .\todoapi.exe
   ```