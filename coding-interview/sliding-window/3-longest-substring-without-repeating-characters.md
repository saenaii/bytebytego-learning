# 3. Longest Substring Without Repeating Characters

## Quick Review

Keep a sliding window that contains no repeated characters. Scan the string with `right` and remember the most recent index of each character.

- When the current character was already seen inside the window, move `left` to one position after its previous index.
- Record the window length `right - left + 1` after updating `left`.
- The scan takes expected `O(n)` time and `O(min(n, alphabet size))` space for the map.

Your submitted solution is correct and already linear. A last-seen-index map can make it simpler: it jumps `left` directly to the next valid position instead of deleting every character passed by `left`.

## Problem

Given a string `s`, return the length of its longest substring that contains no repeated characters.

A substring is contiguous. For example, from `"abcabcbb"`, `"abc"` is a valid substring of length `3`, while `"acb"` is not a substring because its characters are not contiguous.

Examples:

```text
"abcabcbb" -> 3  // "abc"
"bbbbb"    -> 1  // "b"
"pwwkew"   -> 3  // "wke"
""         -> 0
```

## Your accepted solution

```go
func lengthOfLongestSubstring(s string) int {
	res := 0
	existed := make(map[byte]int, len(s))
	for left, right := 0, 0; right < len(s); right++ {
		if idx, ok := existed[s[right]]; ok {
			for left <= idx {
				delete(existed, s[left])
				left++
			}
		}
		existed[s[right]] = right

		res = max(res, right-left+1)
	}
	return res
}
```

Your map represents characters currently inside the window. If the incoming character is already there, the loop deletes characters from the left through its previous occurrence. Then the new character is added at `right`, restoring the no-duplicates rule.

## Can it be improved?

### Performance

The algorithm is already asymptotically optimal. Every character is processed as `right` advances, and `left` only moves forward. Even though the code has an inner `for`, each position is removed at most once, so the total work is expected `O(n)`, not `O(n²)`.

### Simpler pointer movement

The current version deletes map entries while moving `left`. Instead, store each character's most recent index and jump directly past a duplicate. This removes the inner loop and the map deletions:

```go
func lengthOfLongestSubstring(s string) int {
	lastSeen := make(map[byte]int, len(s)) // KEY: character -> most recent index
	left, bestLen := 0, 0                 // KEY: active window start and longest length

	for right := 0; right < len(s); right++ {
		// KEY: only move left when the previous copy is inside the window.
		if previous, ok := lastSeen[s[right]]; ok && previous >= left {
			left = previous + 1 // KEY: exclude the earlier copy
		}
		lastSeen[s[right]] = right // KEY: save this occurrence for future duplicates

		if length := right - left + 1; length > bestLen {
			bestLen = length // KEY: update only after the window is valid
		}
	}

	return bestLen
}
```

The `KEY` comments mark the standard solution's essential state and updates: track previous indexes, keep `left` from moving backward, update the current character's index, and record the best valid window.

This version has the same expected `O(n)` time and `O(min(n, alphabet size))` map space. Its improvement is mainly clarity and fewer map operations, not a better Big-O bound.

The condition `previous >= left` matters. A previous occurrence before the current window must not move `left` backward. For example, in `"abba"`, the last `a` repeats an `a` that is already outside the active window; `left` must stay where it is.

## Window invariant

After processing each character at `right`:

> The current window `s[left:right+1]` contains no repeated characters, and `bestLen` is the longest valid window seen so far.

If `s[right]` was last seen at index `previous` inside the current window, every window that starts at or before `previous` would contain both copies. Moving `left` to `previous + 1` removes the earlier copy while keeping the current window valid.

Then the current valid window has length `right - left + 1`, so it is safe to compare it with `bestLen`.

## Complexity

- **Time:** Expected `O(n)` with a hash map. Each character is processed once; in your submitted version, each position also leaves the window at most once.
- **Space:** `O(min(n, alphabet size))` map entries, which is `O(n)` in the general case.

## Testing without repeated blind attempts

If you need many submissions before a solution passes, use each failing case to test a specific invariant. A short, deliberate test set is often more useful than repeatedly changing code and rerunning the full suite.

| Input | Expected | What it checks |
| --- | ---: | --- |
| `""` | 0 | Empty input |
| `"a"` | 1 | One character |
| `"abc"` | 3 | No repeats |
| `"bbbbb"` | 1 | Repeated adjacent characters |
| `"abba"` | 2 | A repeated character whose old occurrence is outside the active window |
| `"abcabcbb"` | 3 | Repeated pattern; answer is not the whole string |
| `"pwwkew"` | 3 | Repeated characters while the best window is in the middle |

For each test, write down `right`, the previous index, `left` before and after the update, the current window, and its length. Check the invariant after every step: the active window must have no duplicates, and `left` must never move backward.

For extra confidence, compare the optimized function with a simple brute-force solution on many short strings over a tiny alphabet such as `a`, `b`, and `c`. The brute-force version can enumerate every substring and check whether its characters are unique. Small randomized tests quickly expose boundary mistakes while keeping the expected answer easy to verify.

## Assumptions and Go details

The submitted code indexes `s` by byte. That matches the usual LeetCode input constraints for this problem. If the function must treat arbitrary Unicode code points as characters, convert or iterate over runes instead; byte indexing can split a multi-byte character.

The built-in `max` is available in Go 1.21 and later. If using an older Go version, update `res` with a short `if` statement instead.

## Interview answer template

> I’ll keep a window with no repeated characters and scan the string from left to right. A map stores the last index of each character. If a repeated character is still inside the window, I move the left boundary just past its previous occurrence. Then I update the best window length. Since both boundaries only move forward, the scan takes expected `O(n)` time, with space for the character-index map.

## Main takeaways

- A substring must be contiguous, so maintain a valid window rather than selecting arbitrary characters.
- Store last-seen indexes to jump `left` past a duplicate in one update.
- Guard the jump with `previous >= left`; stale occurrences must not move the window backward.
- Your accepted code is already linear. The alternative reduces bookkeeping and makes the pointer rule easier to see.
- When tests keep failing, turn each failure into a small case tied to an invariant, then compare against a brute-force oracle on short inputs.

## Source

[LeetCode 3 — Longest Substring Without Repeating Characters](https://leetcode.com/problems/longest-substring-without-repeating-characters/)
