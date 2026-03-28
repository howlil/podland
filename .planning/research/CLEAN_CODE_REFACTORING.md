# Clean Code Refactoring Guide - Podland Backend

## Research-Based Principles

### 1. DRY (Don't Repeat Yourself)

**Definition:** Every piece of knowledge must have a single, unambiguous, authoritative representation within a system.

**Violations Found:**
- Duplicate SQL queries in `database/queries.go` and `repository/*.go`
- Duplicate entity definitions in `database/types.go` and `entity/*.go`
- Duplicate auth extraction logic in multiple places

**Fix Strategy:**
- Single source of truth for data access: Repository pattern only
- Remove `database/queries.go` entirely
- Keep entities as domain models, database types as persistence models only

---

### 2. KISS (Keep It Simple, Stupid)

**Definition:** Most systems work best if they are kept simple rather than made complicated; therefore, simplicity should be a key goal in design.

**Violations Found:**
- Dual database access patterns (direct SQL + repository)
- Unnecessary `AuthHelper` struct with single-method receivers
- In-memory session storage that won't work in production

**Fix Strategy:**
- Use repository pattern consistently
- Replace helper structs with package-level functions where appropriate
- Use database-backed sessions instead of in-memory

---

### 3. YAGNI (You Ain't Gonna Need It)

**Definition:** Always implement things when you actually need them, never just because you foresee that you might need them.

**Violations Found:**
- Kubernetes integration code not wired into usecase flow
- Unused repository methods (`GetUserByEmail`, `GetAllTiers`, `UpdateUser`)
- Welcome session system that's overly complex for current needs

**Fix Strategy:**
- Remove `internal/k8s/` directory until actually needed
- Remove unused repository methods
- Simplify welcome flow to use database sessions

---

### 4. SOLID Principles

#### Single Responsibility Principle (SRP)
**Definition:** A class should have only one reason to change.

**Current Status:** ✅ Mostly good
- Handlers handle HTTP
- Usecases handle business logic
- Repositories handle data access

**Improvement:**
- Split `handlers/auth.go` into separate concerns (OAuth, session, user management)

#### Open/Closed Principle (OCP)
**Definition:** Software entities should be open for extension but closed for modification.

**Current Status:** ⚠️ Needs work
- Repository interfaces allow extension ✅
- But implementations are tightly coupled to SQL

**Improvement:**
- Keep interfaces, they're good
- Don't over-engineer with mock implementations yet

#### Liskov Substitution Principle (LSP)
**Definition:** Objects of a superclass shall be replaceable with objects of its subclasses.

**Current Status:** ✅ Good
- Repository implementations can be swapped

#### Interface Segregation Principle (ISP)
**Definition:** No client should be forced to depend on methods it does not use.

**Current Status:** ⚠️ Minor issues
- `DB` interface is too large

**Improvement:**
- Keep current interfaces but document which methods are used where

#### Dependency Inversion Principle (DIP)
**Definition:** High-level modules should not depend on low-level modules. Both should depend on abstractions.

**Current Status:** ⚠️ Mixed
- VM handler ✅ uses DI
- Auth handlers ❌ use global `db` variable

**Fix Priority:** HIGH - This is the #1 refactoring needed

---

## Industry Standard Code Patterns

### 1. Dependency Injection Pattern

**Anti-Pattern (Current):**
```go
var db *sql.DB  // Global variable

func HandleLogin(w http.ResponseWriter, r *http.Request) {
    dbWrapper := database.NewDB(db)  // Accessing global
}
```

**Standard Pattern (Target):**
```go
type AuthHandler struct {
    userRepo repository.UserRepository
    sessionRepo repository.SessionRepository
}

func NewAuthHandler(userRepo, sessionRepo) *AuthHandler {
    return &AuthHandler{userRepo, sessionRepo}
}
```

---

### 2. Repository Pattern

**Current Issue:** Dual implementation

**Standard Pattern:**
```go
// Interface in repository/types.go
type UserRepository interface {
    GetUserByID(ctx context.Context, id string) (*entity.User, error)
    // ...
}

// Implementation in repository/user_repository.go
type userRepository struct {
    db *sql.DB
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
    // Uses context for cancellation
}
```

**Remove:** All direct database access from handlers

---

### 3. Context Propagation

**Anti-Pattern (Current):**
```go
// internal/database/queries.go - No context
func (d *sqlDB) GetUserByID(id string) (*User, error) {
    d.db.QueryRow(query, id).Scan(...)  // No cancellation support
}
```

**Standard Pattern:**
```go
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
    r.db.QueryRowContext(ctx, query, id).Scan(...)  // Supports cancellation
}
```

---

### 4. Error Handling

**Current Status:** ✅ Good foundation

**Standard Pattern (already in use):**
```go
// Define sentinel errors
var ErrUserNotFound = errors.New("user not found")

// Wrap with context
return nil, pkgerrors.Wrap(err, "failed to get user")

// Check in handlers
if err == repository.ErrUserNotFound {
    response.NotFound(w, "User not found")
}
```

---

### 5. HTTP Response Standardization

**Anti-Pattern (Current):**
```go
// Some handlers use pkg/response
response.Success(w, http.StatusOK, data)

// Others use direct encoding
json.NewEncoder(w).Encode(data)
```

**Standard Pattern:**
```go
// Always use pkg/response for consistency
response.Success(w, http.StatusOK, data)
response.Error(w, http.StatusBadRequest, "Invalid request")
```

---

## Refactoring Steps (Prioritized)

### Phase 1: Critical Fixes (DRY, DIP violations)

1. **Remove global `db` variable**
   - Create `AuthHandler` with proper DI
   - Inject repositories into auth handlers
   - Update `main.go` to wire dependencies

2. **Remove `internal/database/queries.go`**
   - Move any unique queries to appropriate repositories
   - Update all callers to use repositories
   - Keep only migration code in `database.go`

3. **Consolidate entity/database types**
   - Keep `entity/*` for domain models (no tags)
   - Keep `database/types.go` for persistence models (with tags)
   - Add clear mapping between them in repositories

### Phase 2: Code Organization (KISS, YAGNI)

4. **Merge handler directories**
   - Create `internal/handler/` as single source
   - Move `handler/` and `handlers/` content there
   - Organize by feature: `auth/`, `user/`, `vm/`, `activity/`

5. **Remove unused k8s code**
   - Delete `internal/k8s/` directory
   - Remove k8s dependencies from go.mod (optional)
   - Remove k8s fields from VM entity if not used

6. **Remove unused repository methods**
   - `GetUserByEmail` - not called anywhere
   - `GetAllTiers` - not called anywhere
   - `UpdateUser` - not called anywhere

### Phase 3: Standards Compliance

7. **Add context.Context to all DB operations**
   - Update repository interfaces to include `ctx context.Context`
   - Update all implementations to use `QueryRowContext`, `ExecContext`, etc.
   - Propagate context from handlers → usecases → repositories

8. **Standardize response handling**
   - Replace all `json.NewEncoder` calls with `pkg/response`
   - Ensure consistent response format across all endpoints

9. **Fix in-memory session storage**
   - Replace `welcomeSessions sync.Map` with database-backed sessions
   - Or remove welcome flow entirely if not needed

---

## Code Review Checklist (Per Step)

### For Each Refactoring Step:

**Before:**
- [ ] Identify all files that will be changed
- [ ] Create backup/commit point
- [ ] Document current behavior

**During:**
- [ ] Make small, incremental changes
- [ ] Keep tests passing (if any)
- [ ] Update imports

**After:**
- [ ] Run `go build` - must compile
- [ ] Run `go vet` - no warnings
- [ ] Check `golangci-lint` (if available)
- [ ] Verify no functionality changed (same behavior)
- [ ] Update documentation

---

## Architecture Decision Records (ADRs)

### ADR-001: Repository Pattern Only

**Status:** Accepted

**Context:** Code has dual database access patterns causing DRY violations.

**Decision:** Use repository pattern exclusively. Remove direct database access from handlers.

**Consequences:**
- ✅ Single source of truth for data access
- ✅ Easier to test (can mock repositories)
- ✅ Consistent error handling
- ⚠️ Requires refactoring auth handlers

---

### ADR-002: Context Propagation

**Status:** Accepted

**Context:** Database operations don't support cancellation.

**Decision:** All repository methods must accept `context.Context` as first parameter.

**Consequences:**
- ✅ Supports request timeout/cancellation
- ✅ Better resource cleanup
- ⚠️ Breaking change requires updating all callers

---

### ADR-003: Single Handler Directory

**Status:** Accepted

**Context:** Two handler directories (`handler/` and `handlers/`) cause confusion.

**Decision:** Consolidate to `internal/handler/` organized by feature.

**Consequences:**
- ✅ Clear organization
- ✅ Consistent patterns
- ⚠️ Requires updating all imports

---

## Testing Strategy

### Current State:
- Minimal test coverage (only `health_test.go`, `vm_usecase_test.go`)

### Target State:
- Unit tests for all usecases
- Integration tests for repositories (optional for now)
- Handler tests with mocked usecases

### Test File Organization:
```
internal/
  usecase/
    vm_usecase.go
    vm_usecase_test.go  # Unit tests
  repository/
    user_repository.go
    user_repository_test.go  # Integration tests (optional)
handler/
  auth_handler.go
  auth_handler_test.go  # Handler tests with mocked usecases
```

---

## Metrics for Success

### Code Quality Metrics:

| Metric | Before | Target |
|--------|--------|--------|
| Duplicate code blocks | 15+ | <5 |
| Global variables | 3 | 0 |
| Functions without context | 20+ | 0 |
| Unused exported methods | 6 | 0 |
| Handler directories | 2 | 1 |

### Build Metrics:

| Metric | Before | Target |
|--------|--------|--------|
| `go build` time | ~5s | <5s |
| Binary size | ~50MB | <50MB |
| `go vet` warnings | 0 | 0 |
| `golangci-lint` issues | Unknown | <10 |

---

## Rollback Plan

If any refactoring step causes issues:

1. **Git rollback:**
   ```bash
   git stash  # Save current work
   git checkout <last-good-commit>
   ```

2. **Partial rollback:**
   - Revert specific files using `git checkout HEAD -- <file>`
   - Keep working changes, revert broken ones

3. **Debug approach:**
   - Use `git diff` to identify problematic changes
   - Fix incrementally, test after each change

---

## Next Steps

1. Read this guide completely before starting
2. Follow steps in order (they build on each other)
3. Review code after each step using checklist
4. Update this document if patterns need adjustment
5. Commit after each successful step

**Remember:** The goal is **progress, not perfection**. Small, incremental improvements compound over time.
