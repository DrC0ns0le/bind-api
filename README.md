# DNS Management API

## Overview
This is a simple DNS management service built for managing BIND DNS server configuration files. It provides a RESTful API for creating, updating, and deleting DNS zones and records. The goal is to simplify the management of DNS configuration files by automatically generating the necessary configuration files. It also includes a staging system for previewing the configuration file changes before deployment, and version control via Git for tracking changes. This project also helps in automating the deployment process using Ansible.

## Key Features

### DNS Zone Management
- Complete CRUD operations for DNS zones
- Per-zone configuration management

### DNS Record Management
- Support for all DNS record types
- Bulk record operations
- Advanced filtering and search capabilities
- Automatic PTR record generation (in-addr.arpa, for private IPv4 and IPv6 addresses only)
- Pagination support for large zones
- Support adding tags to records

### Staging System
- Two-stage deployment process (staging → production)
- Preview changes before deployment
- Before/after comparison of zone files
- Version control integration

### Deployment Management
- Automated deployment via Ansible
- Deployment status tracking

### Configuration Management
- System-wide configuration storage
- Staged configuration changes

### Additional Features
- RESTful API design
- Middleware support(CORS, logging etc)
- Soft deletion for all resources
- Configuration via environment variables or flags
- Support for containerization & deployment with Kubernetes
- Uses DB transactions for atomic operations

## To-Do
- Add input validation
- Tagging for records/zones are not tested yet
- Add configuration parse/validation
- Expand on middleware(ie, authentication, metrics etc)
- Write tests
- Add more documentation(API usage, configuration etc)
- Add special tags for automated direct deployments, skipping staging(for purposes of DDNS)

## Technology Stack

- Go standard library net/http and mux
- PostgreSQL for data storage (through pgx package)
- Git for version control integration (through go-git package)
- Ansible for automated deployments

## Getting Started

### Prerequisites

- Go 1.22 or higher (recommended)
- PostgreSQL 13 or higher
- Git
- Ansible (for deployments)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/DrC0ns0le/bind-api.git
```

2. Install dependencies:
```bash
go mod download
```

3. Configure your environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the service:
```bash
go run main.go
```

## API Usage

### Base URL
```
http://your-server:port/api/v1
```

### Example Requests

#### Create a new zone
```bash
curl -X POST http://localhost:8080/api/v1/zones \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com",
    "soa": {
      "primary_ns": "ns1.example.com",
      "admin_email": "admin@example.com",
      "refresh": 3600,
      "retry": 600,
      "expire": 86400,
      "minimum": 3600,
      "ttl": 3600
    }
  }'
```

#### Add a record to a zone
```bash
curl -X POST http://localhost:8080/api/v1/zones/{zone_uuid}/records \
  -H "Content-Type: application/json" \
  -d '{
    "type": "A",
    "host": "www",
    "content": "192.0.2.1",
    "ttl": 3600
  }'
```

## Contributing

Contributions are welcome! You may refer to the TO-DO list above.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This project is provided "as is" without warranty of any kind, either express or implied. The project is developed for personal use and is not intended for production use. Bugs may be present. Use at your own risk.