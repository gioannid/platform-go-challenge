# GWI Platform Go Challenge - Best Practices Implementation

A production-ready REST API for managing user favourites with clean architecture, optional authentication, and extensible storage.

## Features

- **Clean Architecture**: Layered design (domain, repository, service, handler)
- **Thread-Safe In-Memory Storage**: Concurrent-safe with `sync.RWMutex`
- **Optional JWT Authentication**: Enable/disable via environment variable ***(TODO)***
- **Pagination & Sorting**: Efficient handling of large datasets
- **Production Patterns**: Health checks, rate limiting, CORS, recovery middleware ***(TODO)***
- **Extensible Storage**: Easy swap between in-memory and ***(TODO) persisted database***
- **Comprehensive Testing**: Unit and integration tests ***(TODO)***
- **Docker Support**: Dockerfile and docker-compose included ***(TODO)***

## Quick Start

### Using Docker (Recommended)

### Manual Installation

To install and run the application manually, follow these steps:

#### 1. Install Go

**For Windows:**
1.  Visit the official Go website: `https://golang.org/dl/`
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
