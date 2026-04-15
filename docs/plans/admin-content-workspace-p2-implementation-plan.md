# Admin Content Workspace P2 Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Upgrade the current admin article modal into a content-creation workspace with Markdown editing/preview, unified image picking for cover/body, and local draft autosave.

**Architecture:** Keep the active admin mainline on `frontend/admin/index.html + app.js + style.css`. Do not revive `frontend/admin/editor.html` as the primary workflow. Build features incrementally around the existing modal editor and existing image API endpoints.

**Tech Stack:** Vanilla HTML/CSS/JS admin frontend, Go + Gin backend, existing article/category/image endpoints, browser `localStorage`, EasyMDE (preferred reuse path).

---

## 0. Current codebase facts to preserve

### Active frontend files
- `frontend/admin/index.html` — active admin shell and article edit modal
- `frontend/admin/app.js` — active admin behavior, article CRUD, image management, cover preview
- `frontend/admin/style.css` — active admin layout and component styles

### Legacy files
- `frontend/admin/editor.html` — legacy standalone editor page with EasyMDE
- `frontend/admin/images.html` — legacy standalone image page

### Active backend files
- `backend/controllers/article.go` — article CRUD and status handling
- `backend/controllers/upload.go` — image upload/list/delete
- `backend/router/router.go` — API routes
- `backend/models/article.go` — article model with `summary`, `cover_image`, `category_id`, `tags`, `is_top`, `status`

### Important constraints
1. Current main editor is modal-based, not route-based.
2. Current body editor is plain `textarea#editContent`.
3. `openEditor()` already fetches categories and image cache.
4. `saveArticle()` already persists `status`, `cover_image`, `content`, etc.
5. `ListImages` already returns a usable array for picker reuse.
6. No backend autosave endpoint exists; local autosave is the correct P2 approach.

---

# Phase A — Foundation and cleanup

### Task A1: Freeze the active editor mainline

**Objective:** Make the implementation plan explicit in code comments and team practice: all new work goes into the modal editor, not legacy pages.

**Files:**
- Modify: `frontend/admin/index.html`
- Modify: `frontend/admin/app.js`
- Optional docs note: `docs/PRD-admin-content-workspace-p2.md`

**Steps:**
1. Add a short HTML comment near `#editorModal` noting this is the primary article editor workflow.
2. Add a short JS comment above `openEditor()` noting legacy `editor.html` is not the mainline.
3. Do **not** remove legacy files yet.

**Verification:**
- Read file and confirm comments exist.
- No functional behavior changes.

**Commit message:**
```bash
git commit -m "docs(admin): mark modal editor as active main workflow"
```

---

### Task A2: Inventory reusable legacy editor pieces

**Objective:** Reuse only the parts of `editor.html` that are still valuable, mainly EasyMDE bootstrapping ideas.

**Files:**
- Read: `frontend/admin/editor.html`
- Modify later: `frontend/admin/app.js`

**Steps:**
1. Extract the minimal EasyMDE initialization pattern.
2. Ignore old hardcoded API base usage in legacy file.
3. Ignore old page layout and save flow.

**Verification:**
- Document internally which options to reuse: toolbar, preview, value sync.

**Commit:** none required yet.

---

# Phase B — Markdown editor integration

### Task B1: Load EasyMDE assets in active admin page

**Objective:** Make EasyMDE available inside `index.html`.

**Files:**
- Modify: `frontend/admin/index.html`

**Steps:**
1. Add EasyMDE stylesheet in `<head>`.
2. Add EasyMDE script before `/admin/app.js`.
3. Keep existing admin stylesheet intact.

**Implementation notes:**
Use CDN first, matching the already-used dependency style in legacy file.

**Verification:**
- Open admin page in browser.
- Confirm `window.EasyMDE` exists in DevTools console.

**Commit message:**
```bash
git commit -m "feat(admin): load EasyMDE assets in active admin page"
```

---

### Task B2: Introduce editor instance state in `app.js`

**Objective:** Replace raw textarea usage with a managed editor instance.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. Add module-level variables:
   - `let markdownEditor = null;`
   - `let autosaveTimer = null;`
   - `let editorDirty = false;`
2. Add helper functions:
   - `initMarkdownEditor()`
   - `destroyMarkdownEditor()`
   - `getEditorContent()`
   - `setEditorContent(value)`
3. Ensure helpers fall back gracefully if editor init fails.

**Implementation notes:**
- `initMarkdownEditor()` should target `#editContent`.
- If already initialized, avoid double init.
- Set `spellChecker: false` unless needed.

**Verification:**
- Open editor modal twice; ensure only one editor instance exists.
- Confirm content can be read and set correctly.

**Commit message:**
```bash
git commit -m "feat(admin): add managed markdown editor instance helpers"
```

---

### Task B3: Wire editor lifecycle into `openEditor()` and `closeEditor()`

**Objective:** Ensure editor initializes, loads content, and is safely closed with modal lifecycle.

**Files:**
- Modify: `frontend/admin/app.js: openEditor(), closeEditor()`

**Steps:**
1. In `openEditor()`, initialize editor after modal becomes visible.
2. Reset editor content using `setEditorContent('')` instead of raw textarea assignment.
3. When loading article detail, call `setEditorContent(a.content || '')`.
4. In `closeEditor()`, decide whether to keep instance alive or destroy it; choose one consistent behavior.
5. Recommended: keep instance alive while page is loaded, but reset state when modal reopens.

**Verification:**
- New article: editor opens empty.
- Existing article: editor opens with stored Markdown.
- Reopen after closing: no duplicated toolbar or broken layout.

**Commit message:**
```bash
git commit -m "feat(admin): connect markdown editor to modal lifecycle"
```

---

### Task B4: Update save flow to use editor value

**Objective:** Save actual Markdown editor content instead of raw textarea value.

**Files:**
- Modify: `frontend/admin/app.js: saveArticle()`

**Steps:**
1. Replace `document.getElementById('editContent').value` with `getEditorContent()`.
2. Keep existing payload shape unchanged.
3. Preserve existing title validation and success flow.

**Verification:**
- Save a new article with Markdown syntax.
- Reload and reopen article; content should match saved Markdown.

**Commit message:**
```bash
git commit -m "feat(admin): save article content from markdown editor"
```

---

### Task B5: Add preview-oriented styling

**Objective:** Make the editor area visually fit the new admin modal.

**Files:**
- Modify: `frontend/admin/style.css`

**Steps:**
1. Add styles for `.EasyMDEContainer`, `.CodeMirror`, toolbar, preview pane.
2. Ensure modal height and overflow work with a rich editor.
3. Improve mobile behavior so toolbar and content remain usable.

**Verification:**
- Desktop: modal remains scrollable and editor usable.
- Mobile/narrow width: no catastrophic overflow.

**Commit message:**
```bash
git commit -m "style(admin): refine modal layout for markdown editor"
```

---

### Task B6: Add keyboard save shortcut

**Objective:** Allow `Ctrl/Cmd + S` to trigger save.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. Add a keydown listener scoped to the modal/editor context.
2. If `metaKey || ctrlKey` and key is `s`, prevent default and call `saveArticle()`.
3. Guard against duplicate save submissions.

**Verification:**
- In modal, pressing `Ctrl+S` or `Cmd+S` triggers save.
- Outside modal, no unwanted interception.

**Commit message:**
```bash
git commit -m "feat(admin): support keyboard shortcut save in article editor"
```

---

# Phase C — Unified image picker for cover and body

### Task C1: Add picker container markup to active admin page

**Objective:** Provide a reusable modal/drawer shell for image selection.

**Files:**
- Modify: `frontend/admin/index.html`

**Steps:**
1. Add a new hidden picker shell near `#editorModal`, for example `#imagePickerModal`.
2. Include:
   - title area
   - mode label
   - image grid container
   - cancel button
3. Do not preload images in HTML; render dynamically from JS.

**Verification:**
- Picker container exists but stays hidden by default.

**Commit message:**
```bash
git commit -m "feat(admin): add reusable image picker modal shell"
```

---

### Task C2: Add image picker state and open/close functions

**Objective:** Centralize cover/body image insertion behavior.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. Add module-level state:
   - `let imagePickerMode = null; // 'cover' | 'markdown'`
2. Add functions:
   - `openImagePicker(mode)`
   - `closeImagePicker()`
   - `renderImagePicker()`
   - `handleImagePick(url)`
3. Reuse `ensureImageCache()` as the picker data source.

**Behavior:**
- `mode === 'cover'`: write URL into cover input and rerender preview.
- `mode === 'markdown'`: insert Markdown image syntax at current cursor.

**Verification:**
- Picker can open in two modes.
- Image list loads from current image cache.

**Commit message:**
```bash
git commit -m "feat(admin): add unified image picker state and actions"
```

---

### Task C3: Add cover picker trigger in editor UI

**Objective:** Allow selecting a cover from the image library without leaving the editor.

**Files:**
- Modify: `frontend/admin/index.html`
- Modify: `frontend/admin/style.css`

**Steps:**
1. Add a button near `#editCover`, such as “从图床选择”.
2. Wire button to `openImagePicker('cover')`.
3. Keep existing random-cover button in preview card.

**Verification:**
- Clicking the new button opens picker.
- Picking an image fills `#editCover` and updates preview.

**Commit message:**
```bash
git commit -m "feat(admin): allow selecting article cover from image library"
```

---

### Task C4: Add Markdown image insertion action

**Objective:** Allow one-click insertion of image Markdown at cursor position.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. Add an EasyMDE custom toolbar button, e.g. `图床图片`.
2. Toolbar action should call `openImagePicker('markdown')`.
3. Implement `insertMarkdownImage(url, altText = '图片描述')`.
4. Insert syntax at current editor cursor position:
```md
![图片描述](URL)
```
5. Restore focus to editor after insert.

**Verification:**
- Place cursor in middle of content.
- Insert image.
- Markdown syntax appears exactly at cursor location.

**Commit message:**
```bash
git commit -m "feat(admin): support inserting image markdown from picker"
```

---

### Task C5: Add picker styles

**Objective:** Make the image picker feel like part of the admin system.

**Files:**
- Modify: `frontend/admin/style.css`

**Steps:**
1. Add layout and card styles for image picker modal.
2. Reuse existing `.image-grid` visual language where possible.
3. Add selected/hover states and compact URL display.

**Verification:**
- Picker is readable and clickable.
- Cover/body modes are visually distinguishable.

**Commit message:**
```bash
git commit -m "style(admin): style unified image picker modal"
```

---

# Phase D — Local autosave draft system

### Task D1: Add draft key helpers and editor snapshot builder

**Objective:** Create a consistent model for local draft persistence.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. Add helpers:
   - `getDraftStorageKey()`
   - `buildEditorSnapshot()`
   - `saveDraftToLocal()`
   - `loadDraftFromLocal()`
   - `clearDraftFromLocal()`
2. Snapshot should include:
   - `title`
   - `summary`
   - `cover_image`
   - `category_id`
   - `tags`
   - `is_top`
   - `status`
   - `content`
   - `updated_at`
   - `editing_id`

**Key rules:**
- New article: `admin:draft:new`
- Existing article: `admin:draft:article:<id>`

**Verification:**
- Draft object is correctly serialized into `localStorage`.

**Commit message:**
```bash
git commit -m "feat(admin): add local draft persistence helpers"
```

---

### Task D2: Add save status indicator UI

**Objective:** Show users what kind of save state they are in.

**Files:**
- Modify: `frontend/admin/index.html`
- Modify: `frontend/admin/style.css`

**Steps:**
1. Add a status area inside editor modal, for example `#editorSaveState`.
2. Support statuses:
   - 未保存
   - 正在自动保存到本地...
   - 已自动保存到本地 HH:mm:ss
   - 自动保存失败
   - 已保存到服务器
3. Use distinct styles for local vs remote save state.

**Verification:**
- Status text visibly changes during editing and saving.

**Commit message:**
```bash
git commit -m "feat(admin): add editor save status indicator"
```

---

### Task D3: Wire autosave listeners to all relevant fields

**Objective:** Automatically persist edits after user changes.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. Add field change listeners for:
   - `#editTitle`
   - `#editSummary`
   - `#editCover`
   - `#editCategory`
   - `#editTags`
   - `#editIsTop`
   - `#editStatus`
   - Markdown editor change event
2. On change, mark dirty and debounce `saveDraftToLocal()`.
3. Recommended debounce: 2000ms.

**Verification:**
- Editing any field updates local draft after debounce delay.
- Repeated typing does not spam localStorage writes.

**Commit message:**
```bash
git commit -m "feat(admin): autosave editor draft to local storage"
```

---

### Task D4: Restore draft on editor open

**Objective:** Offer safe recovery when local draft exists.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. On `openEditor()`, after loading server content for existing articles, check local draft.
2. If draft exists, compare `updated_at` and prompt user whether to restore.
3. For new article flow, if draft exists, prompt restore before applying random cover fallback.
4. If user accepts, hydrate all fields from draft snapshot.
5. If user declines, leave current values unchanged.

**Verification:**
- New article can restore previous local draft.
- Existing article edit does not silently overwrite server-loaded content.

**Commit message:**
```bash
git commit -m "feat(admin): restore local draft when opening article editor"
```

---

### Task D5: Clear local draft after successful server save

**Objective:** Prevent stale drafts from endlessly reappearing.

**Files:**
- Modify: `frontend/admin/app.js: saveArticle()`

**Steps:**
1. After successful save, clear local draft for current editor context.
2. Set status indicator to `已保存到服务器`.
3. Reset dirty state.

**Verification:**
- Save article successfully.
- Close and reopen editor.
- No stale restore prompt appears unless new local changes exist.

**Commit message:**
```bash
git commit -m "feat(admin): clear local draft after successful article save"
```

---

### Task D6: Guard editor close when unsaved local changes exist

**Objective:** Avoid accidental modal close confusion.

**Files:**
- Modify: `frontend/admin/app.js`

**Steps:**
1. In `closeEditor()`, if there are unsaved local changes, confirm intent.
2. Suggested options via `confirm()` for P2 simplicity:
   - Close and keep local draft
   - Cancel close
3. Do not force remote save on close in P2.

**Verification:**
- Dirty modal close triggers prompt.
- Confirming close keeps local draft intact.

**Commit message:**
```bash
git commit -m "feat(admin): warn before closing dirty article editor"
```

---

# Phase E — Polish and verification

### Task E1: Add optional editor stats

**Objective:** Improve writing feedback with lightweight metrics.

**Files:**
- Modify: `frontend/admin/index.html`
- Modify: `frontend/admin/app.js`
- Modify: `frontend/admin/style.css`

**Steps:**
1. Add a small info row for word count.
2. Optionally add reading-time estimate.
3. Update stats on editor change.

**Verification:**
- Metrics change while typing.

**Commit message:**
```bash
git commit -m "feat(admin): show lightweight writing stats in editor"
```

---

### Task E2: Add backend route tests if route surface changes

**Objective:** Keep route coverage honest if any endpoint changes are introduced.

**Files:**
- Modify only if needed: `backend/router/router_test.go`

**Notes:**
This feature should not require new backend routes if implemented as planned. If no backend change occurs, skip.

---

### Task E3: Manual verification checklist

**Objective:** Validate the full user flow before deployment.

**Verification script:**
1. Login to admin.
2. Open new article modal.
3. Confirm Markdown editor loads.
4. Type content and wait 2 seconds.
5. Confirm local autosave state appears.
6. Close modal.
7. Reopen new article.
8. Confirm restore prompt appears.
9. Accept restore.
10. Use “从图床选择” to set cover.
11. Use Markdown toolbar image button to insert body image.
12. Save article as draft.
13. Reopen article and confirm content persists from server.
14. Edit again, trigger autosave, refresh page, confirm restore flow still works.
15. Publish article and confirm local draft is cleared.

---

# File-by-file implementation summary

## `frontend/admin/index.html`
Planned changes:
- Load EasyMDE CSS/JS
- Add comments clarifying modal editor is mainline
- Add save-state indicator markup
- Add cover picker trigger button
- Add unified image picker modal shell
- Optionally add writing stats container

## `frontend/admin/app.js`
Planned changes:
- Add Markdown editor lifecycle helpers
- Replace direct textarea value reads/writes
- Add keyboard save shortcut
- Add unified image picker state and handlers
- Add Markdown image insertion at cursor
- Add autosave draft helpers and debounce logic
- Add draft restore flow
- Add save state indicator updates
- Add dirty-close confirmation

## `frontend/admin/style.css`
Planned changes:
- Style EasyMDE inside existing modal system
- Improve modal body scrolling/layout
- Add styles for save-state/status bar
- Add styles for unified image picker
- Add styles for optional writing stats

## `backend/controllers/article.go`
Likely no feature changes needed for P2 core scope.

## `backend/controllers/upload.go`
Likely no feature changes needed for P2 core scope.

## `backend/router/router.go`
Likely no feature changes needed for P2 core scope.

---

# Suggested implementation order for real work

1. B1-B4 — get Markdown editor truly working in active modal
2. C1-C4 — unify image picker and connect cover/body insertion
3. D1-D5 — ship autosave persistence and recovery
4. B5, C5, D6, E1 — polish UX
5. E3 — full manual verification

---

# Recommended testing commands

## Frontend/manual
Because admin frontend is static JS/HTML, browser verification is primary.

## Backend tests
From `/root/blog-butterfly-go/backend`:

```bash
go test ./...
```

If route tests were touched:

```bash
go test ./router -v
```

## Full project checks
If this repo has a frontend deployment pipeline or static asset sync step, run the existing deployment verification after manual browser testing.

---

# Rollout notes

1. Keep `editor.html` and `images.html` as legacy fallbacks during implementation.
2. Do not expose a second article editing route.
3. After the new modal workflow is stable, consider a later cleanup PR to retire legacy editor pages.
4. Because the user workflow requires Git commits before deployment, implementation should finish with:

```bash
git status
git add .
git commit -m "feat(admin): upgrade content creation workspace"
git push
```

---

# Completion criteria

This plan is complete when all of the following are true:

- Active admin modal uses Markdown editor instead of plain textarea
- User can preview Markdown
- User can choose a cover directly from image library
- User can insert body images at editor cursor position
- Local autosave draft works for new and existing article editing
- Local draft restore is explicit and safe
- Server save clears local draft
- User can distinguish local autosave from remote persistence
