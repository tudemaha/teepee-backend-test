# 🛒 Teepee Marketplace — Backend

A RESTful marketplace backend built with **Go**, designed as a technical assessment. It implements a full marketplace lifecycle: user authentication, shops, products, cart, orders, payments, and reviews — all backed by clean architecture principles.

---

## 📌 Description

This project simulates a real-world e-commerce backend where:

- **Buyers** can browse products, manage their cart, place orders, pay, and leave reviews
- **Sellers** can open a shop, manage their product catalog (stock, availability, images), and fulfill orders
- **Admins** can manage product categories and oversee platform-wide operations

The project focuses on correctness, consistency, and concurrency — ensuring data integrity even under parallel requests.

---

## 🛠️ Technology Stack

| Category             | Technology                                        |
| -------------------- | ------------------------------------------------- |
| **Language**         | Go 1.26                                           |
| **HTTP Framework**   | Echo v5                                           |
| **ORM**              | GORM v2                                           |
| **Database**         | PostgreSQL                                        |
| **Authentication**   | JWT (Access + Refresh Token) via `golang-jwt/jwt` |
| **Password Hashing** | bcrypt via `golang.org/x/crypto`                  |
| **UUID Generation**  | Google UUID v7                                    |
| **Validation**       | go-playground/validator v10                       |
| **Concurrency**      | `golang.org/x/sync/errgroup`                      |
| **Testing**          | Testify Suite + in-memory SQLite                  |
| **Config**           | godotenv                                          |

---

## 🏗️ Project Structure

```
marketplace-be/
├── cmd/
│   └── main.go                          # Entry point & dependency wiring
├── config/
│   └── config.go                        # Load env vars into typed struct
├── internal/
│   ├── domain/
│   │   ├── entity/                      # GORM DB models
│   │   │   ├── user.go
│   │   │   ├── shop.go
│   │   │   ├── category.go
│   │   │   ├── product.go
│   │   │   ├── cart.go
│   │   │   ├── order.go
│   │   │   ├── order_detail.go
│   │   │   ├── payment.go
│   │   │   ├── review.go
│   │   │   └── refresh_token.go
│   │   └── repository/                  # Repository interfaces (contracts)
│   │       ├── tx_manager.go            # Transaction manager interface
│   │       └── ...
│   ├── usecase/                         # Business logic layer
│   │   ├── auth_usecase.go
│   │   ├── shop_usecase.go
│   │   ├── category_usecase.go
│   │   ├── product_usecase.go
│   │   ├── cart_usecase.go
│   │   ├── order_usecase.go
│   │   ├── payment_usecase.go
│   │   └── review_usecase.go
│   ├── repository/
│   │   └── postgres/                    # GORM implementations
│   │       ├── tx_manager.go            # Transaction manager (RunInTx)
│   │       └── ...
│   ├── delivery/
│   │   └── http/
│   │       ├── dto/                     # Request & Response DTOs
│   │       ├── handler/                 # Echo route handlers
│   │       └── middleware/              # JWT auth & role guard
│   └── tests/
│       ├── setup.go                     # Test server bootstrap (SQLite in-memory)
│       └── e2e_marketplace_test.go      # Full E2E integration test suite
├── pkg/
│   ├── jwt/                             # Token generation & validation
│   ├── password/                        # bcrypt hash & compare
│   ├── response/                        # Unified JSON response helpers
│   └── validator/                       # Echo validator wrapper
├── .env.example
└── go.mod
```

---

## ⚡ Concurrency & Transactions

### 🔀 Goroutines — Parallel DB Reads with `errgroup`

Wherever two or more **independent database lookups** are needed before a write, they are executed **concurrently** using `golang.org/x/sync/errgroup` to reduce latency:

- **`ProductUseCase.Create`** — fetches shop ownership and all categories in parallel
- **`ProductUseCase.Update / UpdateStock / UpdateAvailability`** — fetches product and shop concurrently
- **`ReviewUseCase.CreateReview`** — checks product existence and order completion simultaneously

```go
// Example: concurrent validation in ProductUseCase
g, ctx := errgroup.WithContext(context.Background())

g.Go(func() error {
    shop, err = u.shopRepo.FindByOwnerID(userID)
    return err
})
g.Go(func() error {
    categories, err = u.categoryRepo.FindByIDs(req.CategoryIDs)
    return err
})

if err := g.Wait(); err != nil {
    return nil, err
}
```

> **Rule:** DB _reads_ (validations/lookups) run concurrently. DB _writes_ remain synchronous and always block the HTTP response.

### 🔒 Transactions — Atomic Multi-Table Writes

All operations that touch **multiple tables** are wrapped in a DB transaction using a `TxManager` interface following the **Context Injection Pattern**:

| Use Case                             | What's Atomic                                                                                |
| ------------------------------------ | -------------------------------------------------------------------------------------------- |
| `ShopUseCase.Create`                 | Create shop + upgrade user role to `seller`                                                  |
| `OrderUseCase.Checkout`              | Create order + order details + reduce stock + mark cart checked out + create pending payment |
| `OrderUseCase.UpdateStatus`          | Update order status + restore stock (on cancellation)                                        |
| `PaymentUseCase.UpdatePaymentStatus` | Update payment + trigger order status update                                                 |

```go
// Example: atomic checkout
err = u.txManager.RunInTx(ctx, func(ctx context.Context) error {
    // 1. Create order
    // 2. Create order details (snapshot price)
    // 3. Reduce stock atomically (race-condition safe)
    // 4. Mark cart items as checked out
    // 5. Create pending payment record
    return nil
})
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 14+

### 1. Clone the repository

```bash
git clone https://github.com/tudemaha/teepee-backend-test.git
cd teepee-backend-test
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
APP_PORT=8080
APP_ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=teepee_marketplace
DB_SSLMODE=disable

JWT_SECRET=your-super-secret-key-change-in-production

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=admin123
```

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Run the application

```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080`. The database schema is automatically migrated via GORM `AutoMigrate` on startup. A default admin user is also seeded from the `ADMIN_EMAIL` and `ADMIN_PASSWORD` env vars.

---

## 📡 API Endpoints

| Method   | Endpoint                            | Auth           | Description                             |
| -------- | ----------------------------------- | -------------- | --------------------------------------- |
| `POST`   | `/api/v1/auth/register`             | —              | Register a new user                     |
| `POST`   | `/api/v1/auth/login`                | —              | Login and get token pair                |
| `POST`   | `/api/v1/auth/refresh`              | —              | Rotate access token                     |
| `POST`   | `/api/v1/auth/logout`               | —              | Revoke refresh token                    |
| `GET`    | `/api/v1/auth/me`                   | JWT            | Get own profile                         |
| `POST`   | `/api/v1/shops`                     | JWT            | Create a shop (upgrades user to seller) |
| `GET`    | `/api/v1/shops/:id`                 | —              | Get shop by ID                          |
| `PUT`    | `/api/v1/shops/:id`                 | JWT (owner)    | Update shop info                        |
| `GET`    | `/api/v1/categories`                | —              | List all categories                     |
| `POST`   | `/api/v1/categories`                | Admin          | Create a category                       |
| `DELETE` | `/api/v1/categories/:id`            | Admin          | Delete a category                       |
| `GET`    | `/api/v1/products`                  | —              | List all available products             |
| `GET`    | `/api/v1/products/:id`              | —              | Get product by ID                       |
| `POST`   | `/api/v1/products`                  | Seller         | Create a product                        |
| `PUT`    | `/api/v1/products/:id`              | Seller (owner) | Full product update                     |
| `PATCH`  | `/api/v1/products/:id/stock`        | Seller (owner) | Increment product stock                 |
| `PATCH`  | `/api/v1/products/:id/availability` | Seller (owner) | Toggle product availability             |
| `DELETE` | `/api/v1/products/:id`              | Seller (owner) | Soft delete product                     |
| `GET`    | `/api/v1/products/:id/reviews`      | —              | Get reviews for a product               |
| `GET`    | `/api/v1/carts`                     | JWT            | Get current user's cart                 |
| `POST`   | `/api/v1/carts`                     | JWT            | Add item to cart                        |
| `PUT`    | `/api/v1/carts/:id`                 | JWT            | Update cart item quantity               |
| `DELETE` | `/api/v1/carts/:id`                 | JWT            | Remove item from cart                   |
| `POST`   | `/api/v1/orders/checkout`           | JWT            | Checkout cart (atomic)                  |
| `GET`    | `/api/v1/orders`                    | JWT            | List own orders                         |
| `GET`    | `/api/v1/orders/:id`                | JWT            | Get order details                       |
| `PATCH`  | `/api/v1/orders/:id/status`         | Seller/Admin   | Update order status                     |
| `POST`   | `/api/v1/payments`                  | JWT            | Create a payment for an order           |
| `GET`    | `/api/v1/payments/:orderId`         | JWT            | Get payment by order ID                 |
| `PATCH`  | `/api/v1/payments/:id/status`       | Admin          | Update payment status                   |
| `POST`   | `/api/v1/reviews`                   | JWT (buyer)    | Submit a product review                 |
| `DELETE` | `/api/v1/reviews/:id`               | JWT (owner)    | Delete own review                       |

---

## 🧪 Testing

The project uses a **full E2E integration test suite** built with Testify. Tests run against an **in-memory SQLite database** — no external DB setup required.

The suite simulates a complete real-world marketplace lifecycle across 3 concurrent actors (Buyer, Seller, Admin):

```
Auth → Shop Creation → Category Setup → Product Management →
Cart Operations → Checkout → Payment → Order Completion → Review
```

All **32 implemented endpoints** are explicitly hit and asserted in the test suite.

### Run the tests

```bash
go test -v ./internal/tests
```

### Expected output

```
--- PASS: TestMarketplaceSuite (2.5s)
    --- PASS: TestMarketplaceSuite/Test_01_Auth_RegisterAndLogin
    --- PASS: TestMarketplaceSuite/Test_02_CreateShop
    --- PASS: TestMarketplaceSuite/Test_02b_ShopReadsAndUpdates
    --- PASS: TestMarketplaceSuite/Test_03_Admin_CreateCategory
    --- PASS: TestMarketplaceSuite/Test_03b_CategoryReadsAndDelete
    --- PASS: TestMarketplaceSuite/Test_04_CreateProduct
    --- PASS: TestMarketplaceSuite/Test_04b_ProductUpdatesAndReads
    --- PASS: TestMarketplaceSuite/Test_05_AddToCartAndCheckout
    --- PASS: TestMarketplaceSuite/Test_05b_CartUpdatesAndDelete
    --- PASS: TestMarketplaceSuite/Test_06_PayAndCompleteOrder
    --- PASS: TestMarketplaceSuite/Test_06b_OrderAndPaymentReads
    --- PASS: TestMarketplaceSuite/Test_07_ReviewProduct
    --- PASS: TestMarketplaceSuite/Test_07b_ReviewReadsAndDelete
    --- PASS: TestMarketplaceSuite/Test_08_DeleteProduct
    --- PASS: TestMarketplaceSuite/Test_09_Auth_Me_And_Logout
PASS
ok  	github.com/tudemaha/marketplace-be/internal/tests
```
