# CSUSM GDSC API

The official API for the CSUSM Google Developer Student Club (GDSC) platform. This API handles user authentication, event management, and club-related operations.

## 🚀 Features

- Event Management System
- User Management
- Role-based Access Control
- OAuth2 Authentication (Google & GitHub)
- JWT-based Session Management

## 🛠️ Tech Stack

- Go (1.23+)
- PostgreSQL
- Docker
- Echo Framework
- JWT Authentication
- OAuth2

## 📋 Prerequisites

- [Go](https://go.dev/doc/install) 1.23 or higher
- PostgreSQL
- Docker (optional)

## 🔧 Setup

1. Clone the repository:
```bash
git clone https://github.com/csusmGDSC/csusmgdsc-api.git
cd csusmgdsc-api
```
2. Create and configure your .env file:
```bash
cp .env.example .env
```
3. Install dependencies:
```bash
go mod download
```
4. Run the application:
```bash
go run ./cmd
```

## 📁 Project Structure
```
csusmgdsc-api/
├── cmd/                # Application entrypoint
├── config/             # Configuration management
├── internal/               # Internal application code
│   ├── auth/               # Authentication system
│   ├── handlers/           # Request handlers
│   ├── db/                 # Database operations
│   └── models/             # Database models
├── routes/             # API route definitions
└── .github/workflows/  # CI/CD workflow
```

## 🤝 Contributing
  1. Create your feature branch: `git checkout -b my-new-feature`
  2. Commit your changes: `git commit -m 'Add some feature'`
  3. Push to the branch: `git push origin my-new-feature`
  4. Submit a pull request

### Development Guidelines
    - Follow Go best practices and style guidelines
    - Write tests for new features
    - Update documentation as needed
    - Use meaningful commit messages

## 🔒 Environment Variables
DATABASE_URL=             # PostgreSQL connection string
JWT_ACCESS_SECRET=        # JWT access token secret
JWT_REFRESH_SECRET=       # JWT refresh token secret
GITHUB_CLIENT_ID=         # GitHub OAuth client ID
GITHUB_CLIENT_SECRET=     # GitHub OAuth client secret
GOOGLE_CLIENT_ID=         # Google OAuth client ID
GOOGLE_CLIENT_SECRET=     # Google OAuth client secret
OAUTH_REDIRECT_URL=       # OAuth callback URL
FRONTEND_ORIGIN=          # Frontend application URL

## 🧪 Testing
Run tests: ```go test ./...```

## 👥 Contact
If you have any questions or need help, feel free to reach out to the GDSC team:
- [GDSC Team](https://teams.microsoft.com/l/team/19%3A7u6FOYbIkk7NLclaCv9ucmdDrPBkvXReZm2ixYlEe601%40thread.tacv2/conversations?groupId=8ca48579-37f4-4060-9bf3-cfca2a74f25e&tenantId=128753ab-cb28-4f82-9733-2b9b91d2aca9)
- [CSUSM GDSC Email](mailto:dsccsusm@gmail.com)