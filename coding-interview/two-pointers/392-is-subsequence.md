# 392. Is Subsequence

## Quick Review

This problem uses a **one-way two-pointer scan**:

- `i` tracks the next character to match in `s`.
- `j` scans `t` from left to right.
- Advance `i` only when `s[i] == t[j]`.
- The match succeeds when every character in `s` has been matched.

The key invariant is:

> After processing part of `t`, `i` is the number of characters from `s` that have been matched in order.

The recommended solution takes `O(len(t))` time and `O(1)` extra space. An empty `s` is always a subsequence of `t`.

## Problem

Given two strings `s` and `t`, determine whether `s` is a subsequence of `t`.

A subsequence keeps the original character order, but the matching characters do not need to be adjacent. Characters may be removed from `t`, but they cannot be reordered.

Examples:

```text
s = "abc", t = "ahbgdc" -> true
s = "axc", t = "ahbgdc" -> false
s = "",    t = "anything" -> true
```

## Pattern recognition

This is a two-pointer problem because the task is a one-way matching scan:

- `t` is the larger search space and is scanned from left to right.
- `s` is the target sequence whose order must be preserved.
- A character in `t` that does not match the current character in `s` can be skipped permanently.
- A matching character advances the required position in `s`.

This is not a substring search. A substring must be contiguous; a subsequence only needs to preserve order.

## Round 1: first submission

The first submission used two indexes and returned as soon as it matched the final character of `s`:

```go
ls, lt := len(s), len(t)

var i, j int
for i < ls && j < lt {
	if s[i] != t[j] {
		j++
		continue
	}

	if s[i] == t[j] {
		if i == ls-1 {
			return true
		}
		i++
		j++
	}
}
return false
```

### What works

For a non-empty `s`, the matching logic is sound:

1. If the current characters differ, advance only `j` and keep looking in `t`.
2. If they match, advance both indexes because that character has been used to match the next character in `s`.
3. If the final character of `s` matches, return `true`.
4. If `t` ends first, return `false`.

The algorithm preserves order because `i` and `j` never move backward.

### Why round 1 failed

The empty-string case is missing:

```text
s = "", t = "anything"
```

Here `ls == 0`, so the loop condition `i < ls && j < lt` is false immediately. The function returns `false`, but an empty string is a subsequence of every string. The correct result is `true`.

This is a useful interview lesson: test the smallest valid input before trusting the main loop. Empty inputs often bypass the entire algorithm.

### Other lessons from round 1

- The branch `if s[i] == t[j]` is redundant after the mismatch branch already executes `continue`. Once execution reaches that point, the characters are known to match.
- Checking `i == ls-1` works, but it couples success to a special final-character case. A more general condition is to finish the scan and ask whether all characters in `s` were matched.
- The variable names `ls` and `lt` are compact but less descriptive than `lenS` and `lenT`. For a short interview solution, `i` and `j` are enough if their roles are explained clearly.

## Round 2: successful submission

The second submission added the missing empty-input case:

```go
func isSubsequence(s string, t string) bool {
	ls, lt := len(s), len(t)
	if ls == 0 {
		return true
	}

	var i, j int
	for i < ls && j < lt {
		if s[i] != t[j] {
			j++
			continue
		}

		if s[i] == t[j] {
			if i == ls-1 {
				return true
			}
			i++
			j++
		}
	}
	return false
}
```

This version correctly handles the empty `s` case and is otherwise the same scan as round 1. The second submission fixed the actual failing edge case rather than changing the core algorithm.

## Further optimization

The round 2 solution can be simplified without changing its algorithm. Let `j` scan all of `t`, and let `i` represent how many characters of `s` have already matched:

```go
func isSubsequence(s string, t string) bool {
	i := 0
	for j := 0; j < len(t) && i < len(s); j++ {
		if s[i] == t[j] {
			i++
		}
	}
	return i == len(s)
}
```

### Why this is better

- The empty-string case works naturally: `i == len(s)` is `0 == 0`.
- The redundant equality branch is removed.
- The special `i == ls-1` early return is removed.
- Success is expressed directly as “all characters in `s` were matched.”
- A mismatch only advances `j`; a match advances `i` and then continues scanning `t`.
- The control flow mirrors the invariant, which makes the proof easier to explain.

## Correctness

The optimized solution maintains this invariant:

> After the loop has processed `t[0:j]`, `i` is the number of characters from `s` that have been matched in order using that prefix of `t`.

For each character in `t`:

1. If it does not match `s[i]`, skip it. It cannot help match the current required character later because it has already been passed.
2. If it matches `s[i]`, advance `i`. This consumes the next required character while preserving order.
3. Stop early when `i == len(s)` because every character in `s` has been matched.

If the scan ends before `i` reaches `len(s)`, at least one required character was not found in the correct order, so the result is `false`.

## Complexity

- **Time:** `O(len(t))`. The scan never moves backward and examines each character of `t` at most once.
- **Extra space:** `O(1)`. Only the index `i` and loop variable `j` are used.
- **Best case:** The scan may stop early when all characters in `s` are matched near the beginning of `t`.
- **Worst case:** It scans all of `t`, even when the subsequence does not exist.

The usual constraints for this problem use ASCII or lowercase English letters, so byte indexing is sufficient. A Unicode-aware generalization would need to iterate over runes instead of indexing string bytes.

## Edge cases

| Case | Expected result | Reason |
| --- | --- | --- |
| `s = ""` | `true` | The empty string is a subsequence of every string. |
| `t = ""`, `s != ""` | `false` | There is no character available to match. |
| `len(s) > len(t)` | Usually `false` | `t` cannot provide enough positions for all required characters. |
| Exact match | `true` | Every character matches in order. |
| Characters exist in the wrong order | `false` | A subsequence cannot reorder characters. |
| Repeated characters | Depends on order and count | Each match consumes one position in `s` and one later position in `t`. |

## Lessons from two submissions

- Test empty and minimal inputs before optimizing the main path.
- Separate the essential state transition from redundant branches.
- Prefer a final state check such as `i == len(s)` when it describes success more generally than a special early-return condition.
- A successful submission can still be simplified. Passing is evidence of correctness for tested cases, not proof that the control flow is as clear as it can be.
- State the invariant before writing the loop. It makes both debugging and interview explanation easier.

## Interview answer template

> I’ll scan `t` from left to right with one pointer and keep another index for the next character required from `s`. When the current characters match, I advance the `s` index; otherwise, I only continue scanning `t`. The order is preserved because neither pointer moves backward. At the end, `s` is a subsequence exactly when the number of matched characters equals `len(s)`. The time complexity is `O(len(t))` and the extra space is `O(1)`.
