# 76. Minimum Window Substring

## Quick Review

Use a variable-size window and character frequency counts:

1. Count how many of each character `t` requires.
2. Expand the window with `right`, updating counts for characters that appear in `t`.
3. Once every required character count is satisfied, record the window and move `left` to try a shorter one.
4. Return the shortest valid substring, or `""` if no window works.

Your accepted approach follows this pattern and runs in `O(|s| + |t|)` time with `O(1)` extra space for the fixed ASCII character tables. Its main improvement opportunity is simplifying names and moving the shrink loop out of the branch for required characters. The algorithm itself is already optimal for scanning the input.

## Problem

Given strings `s` and `t`, return the shortest substring of `s` that contains every character in `t`, including repeated characters. If no such substring exists, return an empty string.

For example, if `t` contains two `A`s, a valid window must also contain at least two `A`s. Having the right set of distinct characters is not enough; their required counts must be met.

Example:

```text
s = "ADOBECODEBANC"
t = "ABC"
answer = "BANC"
```

## Your accepted solution

```go
func minWindow(s string, t string) string {
	if len(s) == 0 || len(t) == 0 || len(t) > len(s) {
		return ""
	}

	var need, window [128]int
	counter := 0
	for i := 0; i < len(t); i++ {
		if need[t[i]] == 0 {
			counter++
		}
		need[t[i]]++
	}

	left, valid := 0, 0
	start, length := 0, 0
	for right := 0; right < len(s); right++ {
		if need[s[right]] > 0 {
			window[s[right]]++
			if window[s[right]] == need[s[right]] {
				valid++
			}

			for left <= right && valid == counter {
				if length == 0 || length > right-left+1 {
					start, length = left, right-left+1
				}

				if need[s[left]] > 0 {
					if window[s[left]] == need[s[left]] {
						valid--
					}
					window[s[left]]--
				}
				left++
			}
		}
	}
	return s[start : start+length]
}
```

The algorithm uses two counters with different meanings:

- `counter` is the number of distinct character types required by `t`.
- `valid` is the number of those types whose required frequency is currently met in the window.

When `valid == counter`, the current window contains all required characters with the required counts.

## Can it be improved?

### Algorithm and performance

The sliding-window strategy is already optimal. `right` moves across `s` once, and `left` also only moves forward. The shrink loop is not an extra full scan for every right position: across the whole algorithm, `left` advances at most `len(s)` times.

### Make the control flow flatter

The submitted code runs the shrink loop only when `s[right]` is needed. This is correct: before the next `right` iteration, the previous shrink loop has already stopped because the window is missing a required character. So an irrelevant character cannot make the window valid.

For review, it is a little easier to see the general window rule if the shrink loop sits outside that `if`. We can also give the counters names that state what they count:

```go
func minWindow(s string, t string) string {
	if len(s) == 0 || len(t) == 0 || len(t) > len(s) {
		return ""
	}

	var need, window [128]int
	requiredKinds := 0
	for i := 0; i < len(t); i++ {
		if need[t[i]] == 0 {
			requiredKinds++
		}
		need[t[i]]++
	}

	left, satisfiedKinds := 0, 0
	bestStart, bestLen := 0, 0

	for right := 0; right < len(s); right++ {
		current := s[right]
		if need[current] > 0 {
			window[current]++
			if window[current] == need[current] {
				satisfiedKinds++
			}
		}

		for satisfiedKinds == requiredKinds {
			if bestLen == 0 || right-left+1 < bestLen {
				bestStart, bestLen = left, right-left+1
			}

			leaving := s[left]
			if need[leaving] > 0 {
				if window[leaving] == need[leaving] {
					satisfiedKinds--
				}
				window[leaving]--
			}
			left++
		}
	}

	if bestLen == 0 {
		return ""
	}
	return s[bestStart : bestStart+bestLen]
}
```

The key check is `window[current] == need[current]`: it marks the moment a character type becomes satisfied. If the count grows beyond what is required, it must not increase `satisfiedKinds` again. While shrinking, decrement `satisfiedKinds` only when removing a character would make its count fall below the requirement.

This is a readability refinement, not an asymptotic speedup. The original uses `counter`, `valid`, and `length`; the revised names make their roles easier to remember. Both keep the same fixed-size tables.

## Window invariant

At every step:

> `window[c]` is the count of character `c` in `s[left:right+1]`, and `satisfiedKinds` counts how many distinct required characters currently meet their required counts.

When `satisfiedKinds == requiredKinds`, the window is valid. Record it before removing `s[left]`; after shrinking, the window may become invalid, and the next right-side expansion can make it valid again.

## Why repeated characters matter

Suppose `t = "AABC"`. A window with one `A`, one `B`, and one `C` is still invalid because it is missing the second `A`.

That is why `need` stores a frequency for each character, while `satisfiedKinds` counts distinct character types whose full frequency requirement is met. It does not count matched characters one by one.

## Complexity

- **Time:** `O(|s| + |t|)`. One pass builds `need`; each input position in `s` enters and leaves the window at most once.
- **Extra space:** `O(1)` for the two `[128]int` tables, under the ASCII input assumption.

If using Unicode text rather than the problem's ASCII character set, use rune-aware iteration and maps instead of indexing a 128-entry table by byte.

## Edge cases

| Case | Expected behavior |
| --- | --- |
| `s` or `t` is empty | Return `""`. |
| `len(t) > len(s)` | Return `""`; there are not enough positions to satisfy all required characters. |
| A required character is absent from `s` | Return `""`. |
| `t` contains repeated characters | The window must contain each character at least the required number of times. |
| The shortest window has length one | Return that single character. |
| Several windows have the same minimum length | Return any one of them. |

## Interview answer template

> I’ll count the required frequency of each character in `t`, then scan `s` with a sliding window. I track how many distinct required character types have reached their required counts. When all types are satisfied, I record the current window and shrink from the left until it becomes invalid. Each pointer moves forward at most `|s|` times, so the time is `O(|s| + |t|)`. Under the problem’s ASCII constraints, the frequency tables use `O(1)` space.

## Main takeaways

- Track required **frequencies**, not only whether a character appears.
- Count satisfied distinct character types so duplicate requirements are handled correctly.
- For a minimum valid window, record the answer before shrinking.
- The accepted algorithm is already linear; the useful improvement is clearer counter names and flatter control flow.
- The fixed `[128]int` tables depend on ASCII input.

## Source

[LeetCode 76 — Minimum Window Substring](https://leetcode.com/problems/minimum-window-substring/)
