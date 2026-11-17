# 209. Minimum Size Subarray Sum

## Quick Review

Use a variable-size sliding window to find the shortest contiguous subarray whose sum is at least `target`:

1. Expand the window by adding `nums[right]` to `sum`.
2. While `sum >= target`, record the current length, then remove `nums[left]` and move `left` forward.
3. Return `0` if no valid window was found.

The accepted approach is already optimal: **`O(n)` time and `O(1)` extra space**. Each number enters the sum once and leaves it at most once. Its key assumption is that the array contains positive numbers; with negative numbers, shrinking and expanding no longer change the sum predictably.

## Problem

Given a positive integer `target` and an array of positive integers `nums`, return the minimum length of a contiguous subarray whose sum is greater than or equal to `target`. Return `0` if no such subarray exists.

Example:

```text
target = 7
nums   = [2, 3, 1, 2, 4, 3]
answer = 2  // [4, 3]
```

This is a shortest-window problem. Record a valid window before shrinking it, because every shrink may produce a shorter valid answer.

## Your accepted solution

```go
func minSubArrayLen(target int, nums []int) int {
	left, sum := 0, 0
	result := len(nums) + 1

	for right := 0; right < len(nums); right++ {
		sum += nums[right]

		for sum >= target {
			result = min(result, right-left+1)
			sum -= nums[left]
			left++
		}
	}

	if result > len(nums) {
		return 0
	}
	return result
}
```

The sentinel `len(nums) + 1` is larger than any possible subarray length. If it is still present after the scan, no valid window was found.

## Why the window works

Because every number is positive:

- Adding a number on the right can only increase `sum`.
- Removing a number on the left can only decrease `sum`.

That makes the condition `sum >= target` monotone for this problem. Once the window reaches the target, shrinking is the only way to test a shorter candidate. If shrinking makes the sum too small, the next right-side expansion can make it valid again.

### Window invariant

At the start of each outer-loop iteration, `sum` equals the sum of the current window from `left` through the previous `right`. After adding `nums[right]`, it equals the sum of `[left, right]`.

During the inner loop, the current window is valid. The code records its length before removing the leftmost value, so it cannot miss the shortest valid window for this right endpoint.

## Can it be improved?

### Algorithm and performance

The algorithm is already asymptotically optimal. It runs in `O(n)` time and uses `O(1)` extra space. There is no need for a nested scan: `right` advances through the array once, and `left` also only moves forward.

### Readability

The main style improvement is to rename `result` to `bestLen`, which tells the reader what is being minimized. The `min` call is concise and is supported by Go 1.21 and later. If using an older Go version, replace it with a short `if length < bestLen` check.

```go
func minSubArrayLen(target int, nums []int) int {
	left, sum := 0, 0
	bestLen := len(nums) + 1

	for right := 0; right < len(nums); right++ {
		sum += nums[right]

		for sum >= target {
			bestLen = min(bestLen, right-left+1)
			sum -= nums[left]
			left++
		}
	}

	if bestLen > len(nums) {
		return 0
	}
	return bestLen
}
```

This is a naming improvement, not a faster algorithm. Keeping the final `if` also makes the no-solution case explicit.

## Complexity

- **Time:** `O(n)`. Every element is added once; every element is removed at most once.
- **Extra space:** `O(1)`. The algorithm keeps only indexes, a running sum, and the best length.

The nested `for` and `for sum >= target` loops do not make the algorithm `O(n²)`. Across the whole run, `left` can advance at most `n` times.

## Edge cases

| Case | Expected behavior |
| --- | --- |
| Empty array | Return `0`. |
| No window reaches `target` | Return `0`; the sentinel remains unchanged. |
| One value already reaches `target` | Record a window of length `1`. |
| The shortest valid window ends at the final element | It is recorded before the shrinking step. |
| Several valid windows end at the same `right` | The inner `while` checks them from longest to shortest and records each candidate. |
| Negative values | This sliding-window rule is not valid; use a different technique. |

## Interview answer template

> The answer is a contiguous subarray, and the values are positive, so the running sum changes monotonically as the window expands or shrinks. I’ll add each value from the right. Whenever the sum reaches the target, I record the current length and shrink from the left while it remains valid. Each index enters and leaves the window at most once, so the solution takes `O(n)` time and `O(1)` extra space. If no window qualifies, I return `0`.

## Main takeaways

- For a **minimum-length** window, shrink **while the window is valid** and record the answer before removing an element.
- Positive values make the sum condition monotone, which is why this sliding-window method is correct.
- The accepted algorithm is already optimal; clearer naming is the main code-level refinement.
- A nested `while` can still be linear when both boundaries only move forward.

## Source

[LeetCode 209 — Minimum Size Subarray Sum](https://leetcode.com/problems/minimum-size-subarray-sum/)
