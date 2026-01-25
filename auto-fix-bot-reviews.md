# Auto-Fix Bot Review Comments

You are helping me address automated code review suggestions from `/gemini review` bot on GitHub Pull Requests for the Fluid project. Your goal is to systematically fix **valid** suggestions so PRs can be merged without maintainer back-and-forth.

## Context
- Project: fluid-cloudnative/fluid (Go + Kubernetes)
- Testing: Ginkgo v2 + Gomega
- Go Version: 1.24.12
- My situation: I have ~30 open PRs, maintainer wants bot suggestions addressed before merging

## Your Task

### 1. Analyze Current State
- Check which branch I'm on
- Read modified files: `git diff --name-only origin/master..HEAD`
- Identify bot review suggestions (I'll paste them or you check PR comments)

### 2. Validate Each Suggestion

**Only implement if:**
✅ Real improvement (readability, maintainability, correctness)
✅ Aligns with existing codebase patterns
✅ Won't break functionality

**SKIP if:**
❌ Already implemented (e.g., YAML assertions already exist)
❌ Overly pedantic (e.g., rename with no clarity gain)
❌ Requires major refactoring beyond test migration
❌ Conflicts with existing patterns

**Priority Order:**
1. **HIGH**: Dead code (unused variables, patches, imports)
2. **HIGH**: Magic strings/numbers → constants (if used 2+ times)
3. **HIGH**: Duplicated mock setup (5+ duplications) → Context + BeforeEach
4. **MEDIUM**: BeforeEach for repeated setup (if 3+ test cases)
5. **MEDIUM**: DescribeTable (if 6+ similar test cases with different inputs)
6. **MEDIUM**: Table-driven improvements (suffix logic, reduce duplication)
7. **LOW**: DescribeTable for 2-5 cases, minor renaming
8. **SKIP**: Already done, overly pedantic, or out of scope

**Major refactoring (Context restructuring, large DescribeTable):**
- ✅ DO if it genuinely improves code quality (5+ duplications, 6+ similar tests)
- ❌ SKIP if overly pedantic (2-3 cases) or changes are cosmetic

### 3. Implement Valid Fixes

For each fix:
1. Make the change
2. **Run tests**: `go test -v ./pkg/ddc/[package] -ginkgo.focus="TestName"`
3. Verify it passes
4. If failure → fix or revert

### 4. Commit Strategy

**One commit per logical change:**
```
refactor: extract magic string to constant
test: use BeforeEach for common setup  
chore: remove unused mock patch
```

**Good commit message format:**
```
refactor: extract dataset UID constant

Changed hardcoded "test-dataset-uid" to const dummyDatasetUID
for better maintainability and clarity in shutdown tests.
```

### 5. Final Verification

Before pushing:
- Run all package tests: `go test -v ./pkg/ddc/[package]/...`
- Ensure clean commit history
- No test failures introduced

### 6. Push

```bash
git push -f origin [BRANCH_NAME]
```

---

## Common Valid Patterns (Fix These)

### Pattern 1: Magic Strings → Constants ✅

**When to fix**: String/number used 2+ times OR has semantic meaning

**Before:**
```go
Expect(dataset.Status.UfsTotal).To(Equal("52.25MiB"))
// ... later
dataset.UfsTotal = "52.25MiB"
```

**After:**
```go
const expectedUfsTotal = "52.25MiB"

Expect(dataset.Status.UfsTotal).To(Equal(expectedUfsTotal))
dataset.UfsTotal = expectedUfsTotal
```

### Pattern 2: Repeated Setup → BeforeEach ✅

**When to fix**: Same initialization in 3+ test cases

**Before:**
```go
It("test 1", func() {
    engine := &GooseFSEngine{
        name: "test", 
        namespace: "default",
        Client: fake.NewFakeClient(),
    }
    // test
})
It("test 2", func() {
    engine := &GooseFSEngine{
        name: "test",
        namespace: "default", 
        Client: fake.NewFakeClient(),
    }
    // test
})
```

**After:**
```go
var engine *GooseFSEngine

BeforeEach(func() {
    engine = &GooseFSEngine{
        name: "test",
        namespace: "default",
        Client: fake.NewFakeClient(),
    }
})

It("test 1", func() {
    // test using engine
})
It("test 2", func() {
    // test using engine
})
```

### Pattern 3: Dead Code → Remove ✅

**When to fix**: Always (unused patches, variables, imports)

**Before:**
```go
patch := ApplyFunc(GetReportSummary, func() error { return nil })
defer patch.Reset()
// GetReportSummary never called in test
```

**After:**
```go
// Remove entire patch
```

### Pattern 4: Table-Driven Test Clarity ✅

**When to fix**: Test data has repetitive patterns

**Before:**
```go
Entry("JindoRuntime", "mydata", common.JindoRuntime, "mydata-jindofs-worker"),
Entry("AlluxioRuntime", "data", common.AlluxioRuntime, "data-worker"),
```

**After:**
```go
// Change parameter to just suffix
func(runtimeName, runtimeType, suffix string) {
    Expect(result).To(Equal(runtimeName + suffix))
}
Entry("JindoRuntime", "mydata", common.JindoRuntime, "-jindofs-worker"),
Entry("AlluxioRuntime", "data", common.AlluxioRuntime, "-worker"),
```

---

## Patterns to SKIP

### ❌ Already Implemented

**Bot says**: "Use yaml.Unmarshal for proper assertions"
**Your code**: Already has `yaml.Unmarshal([]byte(result), &config)` at line 683

**Action**: SKIP - already done

### ❌ Overly Pedantic

**Bot says**: "Rename variable `result` to `actualResult`"
**Analysis**: No clarity gain, doesn't match codebase style

**Action**: SKIP - no real improvement

### ❌ Major Refactoring

**Bot says**: "Convert this to DescribeTable with 15 test cases"
**Your change**: Just migrating existing 3 tests to Ginkgo

**Action**: SKIP - out of scope for this PR

---

## Workflow Example

**Me:**
```
Branch: fix/alluxio-cache-tests
PR: #5460

Bot suggests:
1. Extract "test-dataset" to constant
2. Remove unused wrappedUnhookFunc patch  
3. Rename variable for clarity
```

**You analyze:**
1. ✅ Extract constant - used 3 times → **FIX**
2. ✅ Remove patch - never called → **FIX**
3. ❌ Rename - no clarity gain → **SKIP**

**You execute:**
```bash
git checkout fix/alluxio-cache-tests

# Fix 1: Extract constant
# ... make change ...
go test -v ./pkg/ddc/alluxio -ginkgo.focus="cache"
# PASS

git add . && git commit -m "refactor: extract dataset name constant"

# Fix 2: Remove patch
# ... make change ...
go test -v ./pkg/ddc/alluxio -ginkgo.focus="cache"
# PASS

git add . && git commit -m "chore: remove unused wrappedUnhookFunc patch"

git push -f origin fix/alluxio-cache-tests
```

**You report:**
```
✅ Fixed 2/3 suggestions:
- Extracted const testDatasetName (used 3x)
- Removed unused patch

❌ Skipped 1/3:
- Variable rename: no clarity improvement

Tests: PASS
Commits: 2 clean commits  
Push: ✅ Successful
```

---

## Error Handling

- **Test failure**: Debug, fix, or revert change
- **Compilation error**: Fix syntax/imports
- **Unclear suggestion**: Ask me before implementing
- **CI already passing**: Don't over-engineer - keep changes minimal

---

## What You'll Give Me

1. **Changes Made**: List what was fixed (1 line each)
2. **Tests Run**: Commands executed + results  
3. **Commits**: List commit messages
4. **Skipped**: What you didn't fix + why
5. **Push Status**: Success confirmation

---

## Start Command

**Me:**
```
Branch: [branch-name]
PR: #[number]
Bot suggestions: [paste or "check PR"]
```

**You:** Execute the workflow above and report completion.

---

## AUTONOMOUS MODE (All PRs)

**Me:**
```
Follow auto-fix-bot-reviews.md in AUTONOMOUS MODE
```

**You:**
1. Get all open PRs: `gh pr list --author @me --state open --json number,headRefName`
2. For each PR, proactively improve code quality using:
   - Known patterns (constants, dead code, BeforeEach, Context, DescribeTable, scoping)
   - Your own judgment for other improvements (DRY, clarity, Go/Ginkgo best practices)
3. Run tests, commit clearly, push
4. Report progress after each PR
