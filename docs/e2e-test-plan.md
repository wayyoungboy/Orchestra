# Orchestra E2E Test Plan

**Total Tests: 64** | **Created: 2026-04-07**

## Test Categories

### 1. Authentication Tests (8 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| A01 | Login page loads | Navigate to /login, verify page renders | Login form visible with username/password fields |
| A02 | Login with valid credentials | Enter valid username/password, submit | Redirect to /workspaces, auth token stored |
| A03 | Login with invalid credentials | Enter invalid username/password, submit | Error message displayed, stays on login page |
| A04 | Login with empty fields | Submit with empty username/password | Validation error messages shown |
| A05 | Logout functionality | Click logout button | Redirect to /login, token cleared |
| A06 | Auth guard - protected route | Access /workspaces without auth | Redirect to /login |
| A07 | Auth guard - already authenticated | Access /login with valid auth | Redirect to /workspaces |
| A08 | Session persistence | Login, close browser, reopen | Still authenticated (token in localStorage) |

### 2. Workspace Tests (10 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| W01 | Workspace selection page loads | Navigate to /workspaces | Page title, create button, recent workspaces list |
| W02 | Create new workspace button | Click "Create New Workspace" | Modal opens with path browser |
| W03 | Browse server paths | Click path browser, navigate directories | Directory tree visible, selectable |
| W04 | Create workspace with valid path | Select valid path, confirm | New workspace appears in list |
| W05 | Create workspace with invalid path | Enter non-existent path | Error message shown |
| W06 | Select existing workspace | Click workspace in list | Navigate to workspace main view |
| W07 | Remove workspace from list | Click "Remove from list" | Workspace removed, confirmation shown |
| W08 | Workspace switcher dropdown | Click workspace switcher in header | Dropdown shows all workspaces |
| W09 | Switch between workspaces | Select different workspace from dropdown | Active workspace changes, UI updates |
| W10 | Workspace not found | Access non-existent workspace ID | Error page or redirect to /workspaces |

### 3. Chat Tests (12 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| C01 | Chat interface loads | Navigate to workspace chat | Chat header, messages, input visible |
| C02 | Chat sidebar toggle | Click sidebar toggle button | Sidebar expands/collapses |
| C03 | Create conversation | Click "New Conversation" button | Modal opens with conversation options |
| C04 | Send message | Type message, press Enter/send | Message appears in list, WebSocket sends |
| C05 | Receive message (WebSocket) | Simulate incoming message | Message appears in list |
| C06 | Message timestamps | View message list | Each message shows timestamp |
| C07 | Members sidebar | View members sidebar | Member avatars and roles visible |
| C08 | Invite member to chat | Click invite button | Invite modal/link generated |
| C09 | Chat header info | View chat header | Workspace name, member count shown |
| C10 | Chat input validation | Send empty message | Input validation prevents empty send |
| C11 | Conversation selection | Click different conversation in sidebar | Messages switch to selected conversation |
| C12 | Chat scroll behavior | Send multiple messages | Auto-scroll to newest message |

### 4. Terminal Tests (8 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| T01 | Terminal workspace loads | Navigate to /workspace/:id/terminal | Terminal pane visible |
| T02 | Terminal WebSocket connection | Observe terminal WebSocket | Connection established |
| T03 | Terminal input | Type command in terminal | Command executes, output shown |
| T04 | Terminal resize | Resize terminal pane | Terminal adjusts size |
| T05 | Terminal clear | Press clear button (if available) | Terminal content cleared |
| T06 | Terminal session disconnect | Simulate disconnect | Session state preserved or error shown |
| T07 | Terminal multi-pane | Check for multiple terminals | Each pane has independent session |
| T08 | Terminal scroll | Long output, scroll terminal | Scroll works, content preserved |

### 5. Members Tests (10 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| M01 | Members page loads | Navigate to /workspace/:id/members | Members table visible |
| M02 | Add member button | Click "Add Member" button | Add Member modal opens |
| M03 | Add member form validation | Submit with empty fields | Validation errors shown |
| M04 | Add member with valid data | Fill all fields, submit | Member added to table |
| M05 | Member role types | View role dropdown | All roles: owner, admin, secretary, assistant, member |
| M06 | Edit member | Click edit on member row | Edit modal opens with member data |
| M07 | Edit member and save | Change role/name, save | Member updated in table |
| M08 | Delete member | Click delete, confirm | Member removed from table |
| M09 | Member list pagination | Many members (>20) | Pagination controls visible |
| M10 | Member search/filter | Search for member name | Results filtered |

### 6. Settings Tests (6 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| S01 | Settings page loads | Navigate to /workspace/:id/settings | Settings sections visible |
| S02 | Theme toggle | Toggle theme setting | UI theme changes |
| S03 | Workspace settings save | Change setting, save | Settings persisted |
| S04 | API key management | View API key section | Keys visible, can add/remove |
| S05 | Reset to defaults | Click reset defaults button | Settings reset to default values |
| S06 | Settings validation | Invalid settings input | Validation error shown |

### 7. Navigation Tests (6 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| N01 | Sidebar navigation | Click sidebar nav items | Correct page loads |
| N02 | Browser back/forward | Use browser navigation | State restored correctly |
| N03 | Direct URL access | Enter URL directly in browser | Correct page loads |
| N04 | Active nav indicator | View sidebar | Current page highlighted |
| N05 | Home redirect | Access / route | Redirects to /workspaces |
| N06 | Nested route navigation | Navigate within workspace | Child routes work correctly |

### 8. Error Handling Tests (5 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| E01 | API error handling | Simulate API 500 error | Error toast/message shown |
| E02 | WebSocket disconnect | Close WebSocket connection | Reconnection attempt or error shown |
| E03 | Network timeout | Slow network condition | Timeout handling shown |
| E04 | Invalid workspace ID | Access invalid workspace ID | Error page or redirect |
| E05 | Form validation errors | Submit invalid form data | Field-level errors shown |

### 9. WebSocket Tests (4 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| WS01 | WebSocket connection | Navigate to workspace | WebSocket connects |
| WS02 | WebSocket message format | Send/receive message | JSON payload format correct |
| WS03 | WebSocket reconnection | Disconnect, wait | Auto-reconnection |
| WS04 | WebSocket multiple clients | Two browser tabs | Both receive messages |

### 10. UI Responsiveness Tests (5 tests)

| ID | Test Name | Steps | Expected Result |
|----|-----------|-------|-----------------|
| R01 | Mobile viewport | Resize to 375px width | UI adapts to mobile |
| R02 | Tablet viewport | Resize to 768px width | UI adapts to tablet |
| R03 | Large screen | Resize to 1920px width | UI uses full width |
| R04 | Sidebar collapse mobile | Small viewport | Sidebar auto-collapses |
| R05 | Touch interactions | Touch gestures (if mobile) | Touch-friendly UI |

## Test Execution Summary Template

| Category | Total | Passed | Failed | Skipped | Status |
|----------|-------|--------|--------|---------|--------|
| Auth | 8 | - | - | - | - |
| Workspace | 10 | - | - | - | - |
| Chat | 12 | - | - | - | - |
| Terminal | 8 | - | - | - | - |
| Members | 10 | - | - | - | - |
| Settings | 6 | - | - | - | - |
| Navigation | 6 | - | - | - | - |
| Errors | 5 | - | - | - | - |
| WebSocket | 4 | - | - | - | - |
| UI Responsive | 5 | - | - | - | - |
| **TOTAL** | **64** | - | - | - | - |