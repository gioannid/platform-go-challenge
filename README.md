# GWI Platform Go Challenge

A production-ready REST API for managing user favourites with clean architecture, optional authentication, and extensible storage.

## Requirements

Let's say that in GWI platform all of our users have access to a huge list of assets. We want our users to have a peronal list of favourites, meaning assets that favourite or “star” so that they have them in their frontpage dashboard for quick access. An asset can be one the following
* Chart (that has a small title, axes titles and data)
* Insight (a small piece of text that provides some insight into a topic, e.g. "40% of millenials spend more than 3hours on social media daily")
* Audience (which is a series of characteristics, for that exercise lets focus on gender (Male, Female), birth country, age groups, hours spent daily on social media, number of purchases last month)
e.g. Males from 24-35 that spent more than 3 hours on social media daily.

Build a web server which has some endpoint to receive a user id and return a list of all the user’s favourites. Also we want endpoints that would add an asset to favourites, remove it, or edit its description. Assets obviously can share some common attributes (like their description) but they also have completely different structure and data. It’s up to you to decide the structure and we are not looking for something overly complex here (especially for the cases of audiences). There is no need to have/deploy/create an actual database although we would like to discuss about storage options and data representations.

Note that users have no limit on how many assets they want on their favourites so your service will need to provide a reasonable response time.

A working server application with functional API is required, along with a clear readme.md. Useful and passing tests would be also be viewed favourably

It is appreciated, though not required, if a Dockerfile is included.

## Features

- **Clean Architecture**: Layered design (domain, repository, service, handler)
- **Thread-Safe In-Memory Storage**: Concurrent-safe with `sync.RWMutex`
- **Optional JWT Authentication**: Enable/disable via environment variable
- **Pagination & Sorting**: Efficient handling of large datasets
- **Production Patterns**: Health checks, rate limiting etc. middleware ***(TODO)***
- **Extensible Storage**: Easy swap between in-memory and ***(TODO) persisted database***
- **Comprehensive Testing**: Unit and integration tests
- **Docker Support**: Dockerfile and docker-compose included ***(TODO)***

## Installation and Usage

### Using Docker (Recommended)
***(TODO)***

### Manual Installation

To install and run the application manually, follow these steps:

#### 1. Install Go

**For Windows:**
1.  Visit the official Go website: https://golang.org/dl/
2.  Download the MSI installer for your Windows version.
3.  Run the installer and follow the prompts. The installer will automatically add Go to your PATH environment variable.

**For Linux (Debian/Ubuntu):**
1.  Update your package list:
    ```bash
    sudo apt update
    ```
2.  Install Go:
    ```bash
    sudo apt install golang-go
    ```

**For Linux (Fedora/CentOS/RHEL):**
1.  Update your package list:
    ```bash
    sudo dnf update
    ```
2.  Install Go:
    ```bash
    sudo dnf install golang
    ```

**Verify Installation:**
After installation, open a new terminal or command prompt and verify Go is installed correctly:
```bash
go version
```
You should see output similar to `go version go1.22.x <os>/<arch>`.

#### 2. Clone the Repository

Clone the project repository to your local machine:
```bash
git clone https://github.com/gioannid/platform-go-challenge.git
cd platform-go-challenge
```

#### 3. Download Dependencies

Navigate into the project directory and download the required Go modules:
```bash
go mod download
```

#### 4. Run the Server

Start the application server:
```bash
go run cmd/main.go
```
The server will start, by default, on `http://localhost:8080`. You will see a log message indicating the server has started.

**Example output:**
```
2023/10/27 10:00:00 Starting server on :8080
```

You can now access the application's endpoints using a tool like `curl` or Postman. For example, to check the health endpoint:
```bash
curl http://localhost:8080/healthz
```
### Running Tests

To ensure the application's correctness and maintainability, various tests have been implemented. You can run them using the following commands:

*   **Run all tests (unit and integration):**
    ```bash
    go test -v ./...
    ```

*   **Run only unit tests:**
    ```bash
    go test -v ./internal/domain/... ./internal/repository/... ./internal/service/... ./internal/handler/...
    ```

*   **Run only integration tests:**
    ```bash
    go test -v ./test/integration/...
    ```

*   **Run tests with coverage report:**
    ```bash
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    ```
    This will generate an HTML coverage report that you can view in your browser.

*   **Clean test cache:**
    ```bash
    go clean -testcache
    ```

## API Documentation

This project includes interactive Swagger/OpenAPI documentation.

### Accessing Swagger UI

#### 1. Make sure server is running:
```bash
go run cmd/main.go
```

#### 2. Open your browser and navigate to:
http://localhost:8080/swagger/index.html

You can explore all API endpoints, 
view request/response schemas, and even test the API directly from the browser.

### Regenerating Documentation

If you modify API handlers or add new endpoints, regenerate the Swagger docs:

```bash
swag init -g cmd/main.go -o docs
````
