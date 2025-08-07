# BallThai API Server - Project Structure

โปรเจกต์ BallThai API Server ได้รับการจัดระเบียบใหม่เพื่อให้มีโครงสร้างที่ชัดเจนและง่ายต่อการบำรุงรักษา

## โครงสร้างโฟลเดอร์

```
service-ballthai/
├── config/                 # การตั้งค่าระบบ
│   └── config.go           # การจัดการ configuration และ environment variables
├── database/               # Database layer
│   ├── common.go          # database connection และ utilities
│   ├── user.go            # user และ session management
│   ├── coach.go           # coach data operations
│   ├── team.go            # team data operations
│   ├── player.go          # player data operations
│   ├── match.go           # match data operations
│   ├── league.go          # league data operations
│   ├── stadium.go         # stadium data operations
│   ├── standing.go        # standings data operations
│   └── schema.sql         # database schema
├── handlers/               # HTTP handlers (API endpoints)
│   ├── auth.go            # authentication handlers (login, logout, verify)
│   ├── api.go             # main API handlers (leagues, teams, stadiums, matches)
│   └── players.go         # player-specific handlers
├── middleware/             # HTTP middleware
│   └── middleware.go      # CORS และ logging middleware
├── static/                 # Static assets
│   ├── css/
│   │   ├── login.css      # login page styles
│   │   └── dashboard.css  # dashboard page styles
│   └── js/
│       ├── login.js       # login page JavaScript
│       └── dashboard.js   # dashboard page JavaScript
├── templates/              # HTML templates
│   ├── login.html         # login page
│   └── dashboard.html     # dashboard page
├── scraper/               # Data scraping modules
│   ├── api.go
│   ├── coach.go
│   ├── league.go
│   ├── player.go
│   ├── team.go
│   └── ...
├── models/                # Data models (structs)
│   └── ...
├── img/                   # Image storage
│   ├── coach/
│   ├── player/
│   └── stadiums/
├── main_organized.go      # Main application file (organized version)
├── api_auth.go           # Original main file (legacy)
└── go.mod                # Go module dependencies
```

## การใช้งาน

### ไฟล์หลัก
- `main_organized.go` - ไฟล์หลักที่ใช้โครงสร้างใหม่ที่จัดระเบียบแล้ว
- `api_auth.go` - ไฟล์เดิม (legacy) ที่ยังสามารถใช้งานได้

### การรันโปรเจกต์
```bash
# ใช้โครงสร้างใหม่
go run main_organized.go

# หรือใช้ไฟล์เดิม
go run api_auth.go
```

## Features

### 🔐 Authentication System
- Login/Logout ด้วย bcrypt password hashing
- Session management พร้อม token-based authentication
- Protected routes และ middleware

### 📊 API Endpoints
- **Leagues**: `/api/leagues`
- **Teams**: `/api/teams`, `/api/teams/{id}`
- **Players**: `/api/players`, `/api/players/team/{team_id}`, `/api/players/team-post/{team_post_id}`
- **Matches**: `/api/matches`
- **Stadiums**: `/api/stadiums`

### 🌐 Web Interface
- Responsive login page
- Dashboard with API testing capabilities
- Real-time statistics display

### 📱 Frontend Features
- Modern responsive design
- AJAX-based authentication
- Local storage session management
- Error handling และ user feedback

## การตั้งค่า Environment Variables

สร้างไฟล์ `.env` หรือกำหนด environment variables:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=ballthai_db
SERVER_PORT=8080
```

## Dependencies

```bash
go mod init go-ballthai-scraper
go get github.com/gorilla/mux
go get github.com/go-sql-driver/mysql
go get github.com/joho/godotenv
go get golang.org/x/crypto/bcrypt
go get github.com/PuerkitoBio/goquery
go get github.com/rs/cors
```

## การพัฒนาต่อ

### การเพิ่ม Handler ใหม่
1. สร้างฟังก์ชันใน `handlers/` folder ที่เหมาะสม
2. เพิ่ม route ใน `main_organized.go`
3. อัพเดต middleware หากจำเป็น

### การเพิ่ม Model ใหม่
1. สร้าง struct ใน `handlers/` หรือสร้าง `models/` package แยก
2. เพิ่มฟังก์ชัน database operations ใน `database/` folder

### การเพิ่ม Static Assets
1. CSS files → `static/css/`
2. JavaScript files → `static/js/`
3. Images → `static/images/` (ถ้าต้องการ)

## ข้อดีของโครงสร้างใหม่

✅ **การแยกหน้าที่ชัดเจน** - แต่ละ package มีหน้าที่เฉพาะ  
✅ **ง่ายต่อการบำรุงรักษา** - แก้ไขโค้ดตามฟีเจอร์ได้ง่าย  
✅ **Reusable Components** - middleware และ handlers สามารถนำไปใช้ซ้ำได้  
✅ **Better Testing** - แยก logic เป็น functions ทำให้ test ได้ง่าย  
✅ **Scalable** - สามารถเพิ่ม features ใหม่ได้โดยไม่กระทบโค้ดเดิม  

## สถานะปัจจุบัน

- ✅ Authentication system สมบูรณ์
- ✅ API endpoints ทั้งหมดทำงานได้
- ✅ Web interface พร้อมใช้งาน
- ✅ Database schema และ operations
- ✅ Static file serving
- ✅ Project structure reorganization
