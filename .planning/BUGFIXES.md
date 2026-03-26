# Phase 1 Bug Fixes

## Issues Found and Fixed

### Backend Issues

#### 1. Unused Import in `internal/database/queries.go`
**Error:** `"github.com/google/uuid" imported and not used`

**Fix:** Removed unused import.

**File:** `apps/backend/internal/database/queries.go`

---

#### 2. Type Mismatch in `internal/auth/session.go`
**Error:** 
```
cannot use session.DeviceInfo (variable of struct type DeviceInfo) as json.RawMessage value
cannot use oldSession.DeviceInfo (variable of slice type json.RawMessage) as DeviceInfo value
```

**Fix:** Added JSON marshaling/unmarshaling for DeviceInfo when storing to/retrieving from database.

**Changes:**
- Added `encoding/json` import
- Marshal DeviceInfo to JSON before storing: `deviceInfoJSON, _ := json.Marshal(deviceInfo)`
- Unmarshal DeviceInfo from JSON when retrieving: `json.Unmarshal(oldSession.DeviceInfo, &deviceInfo)`

**File:** `apps/backend/internal/auth/session.go`

---

#### 3. Missing Interface Method in `internal/database/types.go`
**Error:** `dbWrapper.UpdateUserNIM undefined (type database.DB has no field or method UpdateUserNIM)`

**Fix:** Added `UpdateUserNIM(userID, nim string) error` to the DB interface.

**File:** `apps/backend/internal/database/types.go`

---

#### 4. Middleware Type Mismatch in `cmd/main.go`
**Error:** 
```
cannot use middleware.AuthMiddleware(handlers.HandleGetMe) (value of interface type http.Handler) 
as func(http.ResponseWriter, *http.Request) value in argument to mux.HandleFunc
```

**Fix:** Changed `AuthMiddleware` signature from `func(http.Handler) http.Handler` to `func(http.HandlerFunc) http.HandlerFunc` to work with `http.HandleFunc`.

**File:** `apps/backend/middleware/middleware.go`

---

### Frontend Issues

#### 1. TanStack Start Complexity
**Issue:** TanStack Start (vinxi) is complex and has many dependencies, causing long install times and potential compatibility issues.

**Fix:** Simplified to use standard Vite + TanStack Router instead of full TanStack Start framework.

**Changes:**
- Removed `@tanstack/start` and `vinxi` dependencies
- Removed `@tanstack/start` specific files (`app.tsx`, `entry-client.tsx`, `entry-server.tsx`, `tanstack.config.ts`)
- Updated `vite.config.ts` with API proxy to backend
- Consolidated all routes into `main.tsx` (simpler for Phase 1)
- Updated package.json scripts

**Files Modified:**
- `apps/frontend/package.json`
- `apps/frontend/vite.config.ts`
- `apps/frontend/src/main.tsx`

---

## Build Verification

### Backend
```bash
cd apps/backend
go build ./...    # ✓ Passes
go test ./...     # ✓ Passes (2 tests)
```

### Frontend
```bash
cd apps/frontend
npm install       # Install dependencies
npm run build     # Build for production
```

---

## Remaining Manual Tests

1. **OAuth Flow** - Requires GitHub OAuth credentials
2. **Database Connection** - Requires PostgreSQL running
3. **Full Integration** - Requires both backend and frontend running

### Test Setup

```bash
# 1. Start PostgreSQL
cd infra/database
docker-compose up -d

# 2. Configure backend
cd apps/backend
cp .env.example .env
# Edit .env with GitHub OAuth credentials

# 3. Start backend
go run ./cmd/main.go

# 4. Start frontend (new terminal)
cd apps/frontend
npm install
npm run dev
```

---

## Test Checklist

- [ ] Backend builds without errors
- [ ] Backend tests pass
- [ ] Frontend installs dependencies
- [ ] Frontend builds without errors
- [ ] OAuth login flow works
- [ ] Student email validation works
- [ ] NIM extraction works
- [ ] Dashboard loads correctly
- [ ] Profile page displays user data
- [ ] Sign out works correctly
- [ ] Session persists after refresh
- [ ] Dark mode works (system preference)
- [ ] Mobile responsive layout works

---

*Bug fixes completed: 2026-03-25*
