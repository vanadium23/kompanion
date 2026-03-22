---
phase: "05"
slug: standalone-koreader-plugin-with-native-highlight-extraction
status: draft
nyquist_compliant: false
wave_0_complete: true
created: 2026-03-22
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Manual / Integration (Lua plugin) |
| **Config file** | none — Lua plugins tested in KOReader |
| **Quick run command** | `go test ./internal/highlight/... -v -short` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~5 seconds |

**Note:** This phase creates a KOReader Lua plugin. Unit testing is manual/integration-based within KOReader environment. The Go tests verify the backend API that the plugin calls.

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/highlight/... -v -short`
- **After every plan wave:** Manual verification in KOReader simulator or device
- **Before `/gsd:verify-work`:** Plugin tested on KOReader device
- **Max feedback latency:** Manual (plugin deployment required)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | D-01, D-02 | manual | N/A — create plugin structure | ⬜ pending | ⬜ pending |
| 05-01-02 | 01 | 1 | D-11, D-12, D-13 | manual | N/A — menu integration | ⬜ pending | ⬜ pending |
| 05-01-03 | 01 | 1 | D-07, D-08, D-09, D-10 | manual | N/A — highlight extraction | ⬜ pending | ⬜ pending |
| 05-01-04 | 01 | 1 | D-04, D-05, D-06 | manual | N/A — HTTP sync logic | ⬜ pending | ⬜ pending |
| 05-01-05 | 01 | 1 | D-14, D-15, D-16 | manual | N/A — error handling | ⬜ pending | ⬜ pending |
| 05-01-06 | 01 | 1 | Integration | integration | `go test ./internal/highlight/... -v` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

**None required.** This phase creates Lua code for KOReader. The existing Go test infrastructure covers the backend API endpoint (`/syncs/highlights`) that the plugin calls.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Plugin appears in Tools menu | D-01 | KOReader UI | Install plugin, open book, check Tools menu |
| Setup dialog saves settings | D-11, D-12, D-13 | KOReader UI | Fill setup dialog, restart KOReader, verify persisted |
| Highlight extraction works | D-07, D-08 | KOReader runtime | Create highlights, run sync, check logs |
| Network handling waits for WiFi | Pattern 3 | KOReader network | Test on device with WiFi off, then enable |
| Success toast shows count | D-14 | KOReader UI | Sync highlights, verify toast message |
| Error toast on auth failure | D-15 | KOReader UI | Use wrong password, verify error message |

---

## Integration Test Flow

1. **Setup:** Install `koreader/kompanion.koplugin/` in KOReader plugins directory
2. **Configure:** Enter Kompanion server URL and device credentials
3. **Create highlights:** Add 3-5 highlights in a test book
4. **Sync:** Tap "Sync highlights" menu item
5. **Verify:** Check Kompanion database for synced highlights

```bash
# Verify highlights in database
psql -d kompanion -c "SELECT COUNT(*) FROM highlights WHERE document_id = '<test_doc_hash>';"
```

---

## Validation Sign-Off

- [x] All tasks have verification method defined
- [x] Manual verifications documented with test instructions
- [x] Wave 0 not needed (Lua plugin, existing Go infra)
- [x] No watch-mode flags
- [ ] Feedback latency: Manual (plugin deployment required)
- [ ] `nyquist_compliant: true` set after manual testing

**Approval:** pending manual testing on KOReader device
