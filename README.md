# iwut-smartclass-backend

## Project Structure

```
.
├── assets
│   └── assets         # Static resource files
│       └── templates  # Templates
├── cmd                # Application entry point
└── internal           # Internal application logic
    ├── application    # Application services
    ├── database       # Database related code
    ├── domain         # Domain logic
    ├── infrastructure # Infrastructure services
    ├── interfaces     # Interface definitions
    └── middleware     # Middleware
```

## Development

### Prerequisites

- Go 1.24+
- MySQL 5.7+

### Build and run

```bash
go build -o server ./cmd
chmod +x ./server
./server
```

### Configuration

**Example:** [.env.example](.env.example)

**Priority:** `Environment variables` > `.env`

## API Documentation

### Get Course Information `POST /getCourse`

**Body:**

```json
{
  "course_name": "高等数学A下",
  "date": "2025-03-26",
  "token": "eyXX"
}
```

**Response:**

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "course_id": 11111,
    "sub_id": 1111111,
    "name": "高等数学A下",
    "teacher": "",
    "location": "",
    "date": "2025-03-26第1-2节",
    "time": "08:00-09:40",
    "video": "https://site/play/default/2025/03/26/123_1920_1080.mp4",
    "summary": {
      "status": "",
      "data": "",
      "model": "deepseek-chat",
      "token": 10000
    }
  }
}
```

### Generate AI Summary `POST /generateSummary`

**Body:**

```json
{
  "sub_id": "1111111",
  "token": "eyXX",
  "task": "new"
}
```

**Response:**

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "sub_id": 1111111,
    "summary_status": "generating"
  }
}
```
