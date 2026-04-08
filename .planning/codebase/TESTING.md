# Testing Patterns

**Analysis Date:** 2026-04-08

## Go Test Framework

**Runner:** `go test` (standard library)
**Config:** No external test config file — uses Go conventions
**Assertion:** Standard library `testing` package only (no testify or third-party assertion library)

**Run Commands:**
```bash
make test              # Run all tests with verbose output
go test -v ./...       # Same as make test
```

### Makefile Test Target

Location: `backend/Makefile`
```makefile
test:
	go test -v ./...
```

### Go Test File Organization

**Location:** Tests are **co-located** with source files using the `*_test.go` naming convention within the same package.

**Detected test files:**
- `backend/internal/ws/gateway_test.go` — WebSocket gateway tests
- `backend/internal/config/loader_test.go` — Config loading tests
- `backend/internal/security/crypto_test.go` — Encryption tests
- `backend/internal/security/whitelist_test.go` — Command/path whitelist tests
- `backend/internal/storage/database_test.go` — Database and migration tests
- `backend/internal/api/middleware/cors_test.go` — CORS middleware tests

### Go Test Patterns

**Simple unit tests** use `t.Fatalf` and `t.Errorf`:
```go
func TestKeyEncryptor_EncryptDecrypt(t *testing.T) {
    key := "test-key-32-bytes-long-1234567890"
    encryptor, err := NewKeyEncryptor(key)
    if err != nil {
        t.Fatalf("NewKeyEncryptor() error = %v", err)
    }
    // ... assertions with t.Errorf()
}
```

**Table-driven tests** using struct slices with named subtests:
```go
func TestIsValidOrigin(t *testing.T) {
    tests := []struct {
        name           string
        origin         string
        allowedOrigins []string
        expected       bool
    }{
        {name: "exact match", origin: "http://localhost:3000", expected: true},
        {name: "no match", origin: "http://malicious.com", expected: false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isValidOrigin(tt.origin, tt.allowedOrigins)
            if result != tt.expected {
                t.Errorf("isValidOrigin(%q, %v) = %v, expected %v", ...)
            }
        })
    }
}
```

**HTTP handler testing** uses `httptest.NewRecorder()` and `httptest.NewRequest()`:
```go
func TestCORS_Middleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(CORS(tt.allowedOrigins))
    r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
    
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    // assert on w.Header() and w.Code
}
```

**Temporary directories** use `t.TempDir()`:
```go
func TestLoadFromFile(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "test.yaml")
    os.WriteFile(configPath, []byte(content), 0644)
    // ... test logic, cleanup is automatic
}
```

**Environment variable tests** use `os.Setenv` / `os.Unsetenv`:
```go
func TestEncryptionKeyFromEnv(t *testing.T) {
    os.Setenv("ORCHESTRA_ENCRYPTION_KEY", "test-key-32-bytes-long-12345678")
    defer os.Unsetenv("ORCHESTRA_ENCRYPTION_KEY")
    // ... test
}
```

**No mocking framework** — tests use real implementations with temporary databases, in-memory structs, or direct instantiation. No gomock, testify/mock, or similar libraries.

## Frontend Testing

### Test Framework

**E2E: Playwright**
- Version: `@playwright/test ^1.58.2`
- Config: `frontend/playwright.config.ts`
- Test directory: `frontend/tests/e2e/` (primary) and `frontend/e2e/` (legacy)

**No unit test framework detected** — no Vitest, Jest, or Vue Test Utils configured.

**Run Commands:**
```bash
pnpm test:e2e              # Build + run Playwright E2E tests
pnpm test:e2e:ui           # Run with Playwright UI
```

### Playwright Configuration

```typescript
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: process.env.ORCHESTRA_E2E_BASE_URL ?? 'http://127.0.0.1:5173',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'pnpm exec vite preview --host 127.0.0.1 --port 5173 --strictPort',
    url: 'http://127.0.0.1:5173',
    reuseExistingServer: true,
    timeout: 60_000,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
```

Key observations:
- Only **Chromium** browser tested (no Firefox, WebKit)
- Uses `webServer` to auto-start Vite preview
- `reuseExistingServer: true` allows manual dev server usage
- CI enables 2 retries
- Traces captured on first retry for debugging

### E2E Test File Organization

**Primary directory:** `frontend/tests/e2e/` (active)
- `chat.spec.ts` — Chat interface and sidebar tests
- `members.spec.ts` — Member management tests
- `errors.spec.ts` — Error handling tests
- `settings.spec.ts` — Settings page tests
- `workspace.spec.ts` — Workspace tests
- `terminal.spec.ts` — Terminal tests
- `message-push.spec.ts` — Message push tests
- `responsive.spec.ts` — Responsive layout tests

**Legacy directory:** `frontend/e2e/` (likely deprecated but still present)
- `smoke.spec.ts` — Basic smoke tests
- `auth-flow.spec.ts` — Authentication flow
- `workspaces-api.spec.ts` — API-level workspace tests
- `ws-push.spec.ts` — WebSocket push tests
- `direct-ws.spec.ts` — Direct WebSocket tests
- `full-message-flow.spec.ts` — End-to-end message flow
- `vue-ws-integration.spec.ts` — Vue-WebSocket integration
- `real-ws-test.spec.ts` — Real WebSocket tests
- `mention.spec.ts` — @mention feature tests

### E2E Test Structure

**Basic test pattern:**
```typescript
import { test, expect } from '@playwright/test'

test.describe('Chat Interface', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/workspaces')
    await page.waitForLoadState('networkidle')
    const workspaceCard = page.locator('.workspace-item-card').first()
    if (await workspaceCard.isVisible()) {
      await workspaceCard.click()
      await page.waitForURL(/\/workspace\/.+\/chat/)
    }
  })

  test('C01: should load chat interface', async ({ page }) => {
    await expect(page.locator('h1')).toBeVisible()
    await expect(page.getByPlaceholder('发送到')).toBeVisible()
  })
})
```

**API interception for error testing:**
```typescript
test('E01: should handle API errors gracefully', async ({ page }) => {
  await page.route('**/api/workspaces', route => {
    route.fulfill({ status: 500, body: '{"error":"Internal server error"}' })
  })
  await page.goto('/workspaces')
  await expect(page.locator('body')).toBeVisible()
})
```

**Backend API testing** (no page needed):
```typescript
test('backend health', async ({ request }) => {
  const api = process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:8080'
  const result = await request.get(`${api}/health`)
  expect(result.ok()).toBeTruthy()
  const body = await result.json()
  expect(body).toHaveProperty('status', 'ok')
})
```

**Test ID naming:** Tests use a `letter-number` prefix for traceability (e.g., `C01`, `M02`, `E04`, `S03`).

### E2E Test Utilities

**No dedicated test utilities or fixtures** detected. Tests:
- Navigate to `/workspaces` and click the first `.workspace-item-card` to reach chat
- Use `page.locator()` with CSS classes and text content
- Use `page.getByRole()` for accessibility-based selectors
- No page object pattern or POM detected
- No global setup/teardown fixtures

**Locators used:**
- CSS classes: `.workspace-item-card`, `.chat-interface-root`
- Text content: `page.locator('text=团队成员')`, `page.getByRole('button', { name: '添加成员' })`
- Placeholders: `page.getByPlaceholder('发送到')`
- Role-based: `page.getByRole('button', { name: '...' })`

### Mocking (Frontend)

**No mocking framework** — Playwright tests use:
- `page.route()` for intercepting and stubbing API calls
- Direct backend calls for integration tests
- `test.skip()` for conditional test skipping when backend unavailable

## Test Coverage

### What IS Tested

**Go backend:**
- Config loading and defaults (`config/loader_test.go`)
- Encryption encrypt/decrypt (`security/crypto_test.go`)
- Command and path whitelisting (`security/whitelist_test.go`)
- Database creation and migration (`storage/database_test.go`)
- CORS middleware origin validation (`api/middleware/cors_test.go`)
- WebSocket gateway creation and origin checking (`ws/gateway_test.go`)

**Frontend E2E:**
- Chat interface loading and message sending
- Members page with add/role management
- Error handling for invalid workspaces
- Settings page navigation
- Responsive layouts
- Terminal availability

### Test Coverage Gaps

**Go backend (no tests detected):**
- **All HTTP handlers** — `handlers/auth.go`, `handlers/workspace.go`, `handlers/member.go`, `handlers/conversation.go`, `handlers/terminal.go` have zero unit test coverage
- **Router setup** — `api/router.go` has no test for route registration
- **All repository implementations** — No tests for `repository/workspace.go`, `repository/member.go`, `repository/message.go`, etc.
- **WebSocket chat gateway** — `ws/gateway.go` HandleChat has no test
- **A2A protocol layer** — `internal/a2a/` has no tests
- **Filesystem validator and browser** — `internal/filesystem/` has no tests
- **Middleware auth** — `middleware/auth.go` has no test (only CORS is tested)

**Frontend (no unit tests):**
- **No component unit tests** — No Vitest or Vue Test Utils configured
- **No store unit tests** — Pinia stores (`chatStore.ts`, `authStore.ts`) untested in isolation
- **No API client tests** — `client.ts` interceptor logic not tested
- **E2E test dependency** — Tests require a running backend; no mock server for frontend-only testing

### Coverage Enforcement

- **No coverage threshold** configured in Makefile or Playwright config
- **No CI pipeline** detected with coverage reporting
- **No `go test -cover`** flag used in `make test`

## Recommendations

1. **Add handler tests** — Critical HTTP handlers have zero coverage; start with auth and workspace handlers
2. **Add `-cover` flag** — Update `make test` to `go test -v -cover ./...` for basic coverage visibility
3. **Add frontend unit tests** — Consider Vitest for Pinia store and component testing
4. **Consolidate E2E directories** — Both `tests/e2e/` and `e2e/` contain Playwright tests; one is likely legacy
5. **Add page object pattern** — E2E tests repeat navigation setup; extract common helpers
6. **Add database repository tests** — SQL layer has no test coverage

---

*Testing analysis: 2026-04-08*
