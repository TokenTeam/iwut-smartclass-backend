# iwut-smartclass-backend

## Project Structure

```
.
├── assets          # Static resource files
├── cmd             # Application entry point
├── internal        # Internal application logic
│   ├── config      # Configuration files and loading logic
│   ├── database    # Database related code
│   ├── handler     # HTTP handlers
│   ├── middleware  # Middleware
│   ├── router      # Route definitions
│   ├── service     # Business logic
│   │   └── course  # Course related services
│   └── util        # Utilities
```

## Development

### Prerequisites

- Go 1.23+
- MySQL 5.7+

### Build

```bash
go build -o server ./cmd
```

## API Documentation

### Get Course Information `GET /getCourse`

Body:

```json
{
  "course_name": "高等数学A下",
  "date": "2025-03-26",
  "token": "eyXX"
}
```

Response:

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
      "data": "",
      "status": ""
    }
  }
}
```
