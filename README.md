# GophKeeper - Secure Secrets Manager

GophKeeper Architecture

GophKeeper is a secure CLI-based secrets manager that allows users to store and manage sensitive information with end-to-end encryption.
The application follows a client-server architecture where sensitive data is encrypted on the client before being stored on the server.

## Key Features

    🔒 End-to-End Encryption: Data encrypted on client before transmission

    🔑 Multi-Device Support: Access your secrets from multiple devices

    🛡️ Zero-Knowledge Architecture: Server never sees unencrypted data

    📱 Cross-Platform: Works on Windows, macOS, and Linux

    📦 Multiple Data Types: Store passwords, cards, notes, and binary data

    🔍 Secure Search: Find secrets without exposing their contents

## Technology Stack
### Client
- Language: Go
- Encryption: RSA-2048, XChaCha20-Poly1305
- Storage: Local encrypted files
- CLI Framework: Cobra

### Server
- Language: Go
- Database: PostgreSQL
- API: gRPC with TLS
- Authentication: JWT tokens

## Installation
### Server Setup

``` bash
# Clone the repository
git clone https://github.com/frolmr/GophKeeper.git
cd GophKeeper/internal/server

# Set required environment variables
export DATABASE_URI=postgres://user:password@localhost/gophkeeper
export TLS_CERT_FILE=path/to/cert.pem
export TLS_KEY_FILE=path/to/key.pem
export JWT_SECRET=$(openssl rand -base64 32)

# Build and run
go build -o gophkeeper-server
./gophkeeper-server
```

### Client Installation

``` bash
# Install the client
cd internal/client
go install

# Configure server address
export GOPHKEEPER_SERVER=your-server-address:50051

# Start using GophKeeper
gophkeeper --help
```


## Usage Examples
### User Registration
``` bash
gophkeeper user register --email user@example.com --password strongpassword
```

### Device Registration
``` bash
gophkeeper device add
```

### Store a Password
``` bash
gophkeeper record add --name "My Email" --kind password
# Follow prompts to enter credentials
```

### List All Records

``` bash
gophkeeper record list
```

### Retrieve a Secret

``` bash
gophkeeper record read <record-uuid>
```

## Security Model

GophKeeper implements a robust security architecture:

1. Master Key Encryption:
- Generated during registration
- Encrypted with device public key
- Stored on server in encrypted form

2. Record Encryption:
- Each record encrypted with master key
- Master key decrypted locally with device private key
- Server never has access to decrypted data

3. Device Verification:
- New devices require confirmation
- Master key re-encrypted for each device
- Device revocation support

## Configuration
### Client Configuration
|Environment Variable  |Description             |Default    |
|----------------------|------------------------|-----------|
|GOPHKEEPER_SERVER	   |Server address	        |-          |
|GOPHKEEPER_CONFIG_DIR |Custom config directory |OS default |

#### Server Configuration
|Environment Variable |Description	                |Default |
|---------------------|-----------------------------|--------|
|DATABASE_URI	      |PostgreSQL connection string |-       |
|RUN_ADDRESS	      |Server bind address          |:50051  |
|TLS_CERT_FILE	      |TLS certificate path	    |-       |
|TLS_KEY_FILE	      |TLS private key path	    |-       |
|JWT_SECRET	      |JWT signing key	            |Random  |


## Architecture Overview

``` bash
Client Components:
├── CLI Interface
├── Cryptographic Engine
├── Local Storage
└── gRPC Client Adapter

Server Components:
├── gRPC API Server
├── Authentication Service
├── Device Management
├── Record Storage
└── PostgreSQL Database
```
