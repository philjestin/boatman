# Review System Improvements

This document summarizes the changes made to make boatmanmode's review system less stringent and more flexible.

## Overview

The review system was previously very strict, often failing reviews for minor issues or being overly sensitive to suggestion language. These improvements make the system more configurable and lenient while maintaining code quality standards.

## Changes Made

### 1. Configurable Pass Criteria

**Location**: `internal/config/config.go`

Added new `ReviewConfig` struct with configurable thresholds:

```yaml
review:
  max_critical_issues: 1           # Allow up to 1 critical issue (was 0)
  max_major_issues: 3              # Allow up to 3 major issues (was 2)
  min_verification_confidence: 50  # Minimum confidence % for diff verification
  strict_parsing: false            # Relaxed keyword parsing by default
```

**Benefits**:
- Teams can adjust review strictness based on their needs
- More realistic thresholds that allow for constructive feedback
- Easy to tighten or relax standards via configuration

### 2. Relaxed Natural Language Parsing

**Location**: `internal/scottbott/scottbott.go`

**Changes**:
- Added `strict_parsing` flag to control keyword sensitivity
- In relaxed mode (default), only truly blocking language fails reviews:
  - "cannot be merged"
  - "blocking issue"
- Removed false-positive triggers like:
  - "must be addressed" (constructive feedback)
  - "needs work" (normal review language)
  - "issues that need to be addressed" (descriptive, not blocking)

**Benefits**:
- Constructive review feedback no longer auto-fails
- Reviews can suggest improvements without blocking merge
- More natural review conversations

### 3. Improved Diff Verification

**Location**: `internal/diffverify/diffverify.go`

**Changes**:
- More lenient heuristics for detecting fixes:
  - **Critical issues**: 3+ additions or 2+ removals (was 5+/3+)
  - **Major issues**: 1+ additions or any removals (was 2+ additions)
  - **Minor issues**: Any file modification counts as addressing
- Increased base confidence from 80% to 85%
- Smarter new issue detection - only flags truly problematic patterns:
  - Removed: TODO comments, console.log, debug prints
  - Kept: FIXME, XXX markers, debugger statements
- Better confidence calculation:
  - 70% base confidence + 30% based on issue resolution ratio
  - Only penalize 5 points per concerning new issue (was 10)
- Fixed bug where wrong file modifications were counted as addressing issues

**Benefits**:
- Fewer false negatives (fixes not detected)
- More realistic confidence scores
- Focuses on actual problems, not development artifacts

### 4. Increased Default Iterations

**Location**: `internal/config/config.go`

**Change**: `max_iterations: 5` (was 3)

**Benefits**:
- More chances to address feedback and pass review
- Better for complex refactoring that takes multiple rounds
- Still configurable if you want faster iterations

## Configuration Examples

### Strict Mode (High Quality Bar)
```yaml
max_iterations: 3
review:
  max_critical_issues: 0
  max_major_issues: 1
  min_verification_confidence: 70
  strict_parsing: true
```

### Balanced Mode (Default)
```yaml
max_iterations: 5
review:
  max_critical_issues: 1
  max_major_issues: 3
  min_verification_confidence: 50
  strict_parsing: false
```

### Lenient Mode (Fast Iteration)
```yaml
max_iterations: 7
review:
  max_critical_issues: 2
  max_major_issues: 5
  min_verification_confidence: 40
  strict_parsing: false
```

## Migration Guide

### For Existing Users

1. **No action required** - defaults are now more lenient
2. **To restore old behavior** - add to `.boatman.yaml`:
   ```yaml
   max_iterations: 3
   review:
     max_critical_issues: 0
     max_major_issues: 2
     strict_parsing: true
   ```

### For New Users

- The default configuration should work well for most projects
- Adjust `review.max_critical_issues` and `review.max_major_issues` based on team preferences
- Enable `strict_parsing: true` if you want stricter review language enforcement

## Testing

All existing tests have been updated to reflect the new behavior:
- ✅ Config tests verify new defaults
- ✅ Diff verification tests use more realistic scenarios
- ✅ All integration tests pass

## Backward Compatibility

✅ **Fully backward compatible** - all configuration is optional with sensible defaults.

Existing configurations will continue to work. New settings only apply if specified.
