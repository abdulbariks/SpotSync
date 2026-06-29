# SpotSync – Smart Parking & EV Charging Reservation

Live URL: TBD (deploy to Render/Railway/Fly.io)

A centralized platform for busy airports and malls to manage parking zones, specifically handling the high-demand reservation of limited EV charging spots.

## Features

- User registration and authentication with JWT
- Role-based access control (driver/admin)
- Parking zone management (CRUD)
- Real-time availability tracking
- Concurrent-safe reservation system with row-level locking
- License plate duplicate check for active reservations

## Tech Stack

| Technology | Version/Notes |
| --- | --- |
| Go (Golang) | 1.22+ |
| Echo | `github.com/labstack/echo/v4` - Web framework |
| GORM | `gorm.io/gorm` - ORM with PostgreSQL driver |
| PostgreSQL | Database (NeonDB/Supabase/Aiven) |
| Validator | `github.com/go-playground/validator/v10` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| bcrypt | `golang.org/x/crypto/bcrypt` |

## Architecture

Clean Architecture with strict layer separation:

```
main.go
  ├── handler/     # HTTP layer - validates DTOs, extracts JWT, calls Service
  ├── service/     # Business logic - password hashing, JWT generation, rules
  ├── repository/  # Data access - GORM operations with transactions
  ├── models/      # GORM structs
  └── dto/         # Request/Response structures
```

## Setup

1. Clone the repository
2. Install Go 1.22+
3. Set up PostgreSQL database (NeonDB, Supabase, or Aiven)
4. Configure `.env`:

```env
DATABASE_URL=postgresql://user:password@host:5432/dbname?sslmode=require
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=24
SERVER_PORT=8080
```

5. Run the server:

```bash
go run main.go
```

## API Endpoints

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | Public | Register new user |
| POST | `/api/v1/auth/login` | Public | Login and get JWT |
| GET | `/api/v1/zones` | Public | List all parking zones |
| GET | `/api/v1/zones/:id` | Public | Get single zone |
| POST | `/api/v1/zones` | Admin | Create zone |
| PUT | `/api/v1/zones/:id` | Admin | Update zone |
| DELETE | `/api/v1/zones/:id` | Admin | Delete zone |
| POST | `/api/v1/reservations` | Auth | Create reservation |
| GET | `/api/v1/reservations/my-reservations` | Auth | Get user's reservations |
| DELETE | `/api/v1/reservations/:id` | Auth | Cancel reservation |
| GET | `/api/v1/reservations` | Admin | List all reservations |

## Concurrency Handling

Reservations use GORM database transactions with row-level locking (`FOR UPDATE`) to prevent race conditions when checking capacity and creating reservations atomically.