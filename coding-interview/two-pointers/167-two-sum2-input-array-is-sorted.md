# 167. Two Sum II — Input Array Is Sorted

## Quick Review

Use two pointers at the ends of the sorted array:

- `left` starts at the first element.
- `right` starts at the last element.
- If the sum is too small, move `left` right to increase it.
- If the sum is too large, move `right` left to decrease it.
- Return the 1-based indexes when the sum equals `target`.

The sorted order makes each pointer movement safe. The recommended solution runs in `O(n)` time and uses `O(1)` extra space.

## Problem

Given a 1-indexed array of integers `numbers` sorted in non-decreasing order, find two numbers whose sum equals `target`.

Return their 1-based indexes as `[leftIndex, rightIndex]`. The problem guarantees that exactly one valid answer exists, and the same element cannot be used twice.

Example:

```text
numbers = [2, 7, 11, 15], target = 9
answer = [1, 2]
```

## Pattern recognition

This is an inward two-pointer problem because:

- The input is sorted.
- The task asks for a pair of values.
- The current sum tells us which side cannot produce a solution with the current opposite pointer.

The sorted order is the key property. Without it, moving one pointer based on the current sum would not safely eliminate candidates.

## Submitted approach

The submitted solution starts with the correct pointer positions and uses the sum to move the pointers:

```go
func twoSum(numbers []int, target int) []int {
	left, right := 0, len(numbers)-1
	for left < right {
		if numbers[left]+numbers[right] == target {
			return []int{left + 1, right + 1}
		}

		for left < right && numbers[left]+numbers[right] > target {
			right--
		}
		for left < right && numbers[left]+numbers[right] < target {
			left++
		}
	}
	return []int{left + 1, right + 1}
}
```

### What works

- It correctly uses the 1-based indexes required by the problem.
- It starts with the smallest and largest values, which gives the widest possible sum range.
- When the sum is too large, moving `right` left cannot skip a valid pair with the current `left`: every earlier value is no larger and would make the sum even smaller.
- When the sum is too small, moving `left` right cannot skip a valid pair with the current `right`: every later value is no smaller and would make the sum larger.
- Because both pointers only move inward, the total pointer movement remains linear.

## What can be improved

### 1. The inner loops are unnecessary

The two inner loops repeatedly recalculate the same sum and move a pointer several positions in one outer iteration. This can still work under the problem’s guarantee, but it makes the control flow harder to read and prove.

The standard two-pointer transition is simpler:

```text
if sum == target: return the pair
if sum < target: left++
if sum > target: right--
```

Only one pointer needs to move per iteration because the current comparison already tells us which direction can improve the sum.

### 2. The fallback return is unsafe outside the problem guarantee

The submitted code ends with:

```go
return []int{left + 1, right + 1}
```

If no pair exists, the pointers may have met or crossed, so this can return an invalid pair such as the same index twice or indexes that do not represent a solution. The problem guarantees one answer, so this branch is unreachable for valid judge inputs, but `return nil` is a safer general-purpose fallback.

### 3. Compute the sum once per iteration

The submitted code evaluates `numbers[left] + numbers[right]` repeatedly inside nested conditions. Store it in a local variable so the comparison and the pointer movement describe one clear decision.

## Recommended solution

```go
func twoSum(numbers []int, target int) []int {
	left, right := 0, len(numbers)-1

	for left < right {
		sum := numbers[left] + numbers[right]

		switch {
		case sum == target:
			return []int{left + 1, right + 1}
		case sum < target:
			left++
		default:
			right--
		}
	}

	return nil
}
```

This version keeps the same algorithmic idea but makes the decision tree explicit and removes repeated work.

## Correctness invariant

Before every iteration:

> If a valid pair still exists, it must be within the inclusive range `[left, right]`.

The invariant is preserved as follows:

- If `sum < target`, every pair using `numbers[left]` with an index at or before `right` is too small. Since the array is sorted, discard `left` by moving it right.
- If `sum > target`, every pair using `numbers[right]` with an index at or after `left` is too large. Discard `right` by moving it left.
- If `sum == target`, the current pair is the unique answer guaranteed by the problem.

When the pointers meet, no two distinct indexes remain, so no valid pair exists. Under the problem guarantee, the loop returns before that happens.

## Complexity

- **Time:** `O(n)`. Each pointer moves inward at most `n` times, and the array is already sorted.
- **Extra space:** `O(1)`, excluding the returned two-element slice.
- **No sorting cost:** Sorting is unnecessary and would add `O(n log n)` time while potentially losing the original index positions.

## Edge cases

| Case | Expected handling |
| --- | --- |
| Two elements whose sum is the target | Return `[1, 2]`. |
| Target equals the smallest possible pair sum | The first comparison can succeed immediately. |
| Target equals the largest possible pair sum | The first comparison can succeed immediately. |
| Duplicate values | Move both pointers only when their sum matches; duplicates are valid if they occupy different indexes. |
| Negative values | The same comparison rules work because sorted order, not positivity, drives pointer movement. |
| No pair exists | Return `nil` in general-purpose code; the problem’s constraints guarantee this case does not occur. |
| One-element or empty input | The loop does not run; return `nil` unless the caller guarantees a valid pair. |

## Interview answer template

> The array is sorted, so I’ll place one pointer at each end. If their sum is smaller than the target, I move the left pointer right because that is the only direction that can increase the sum. If the sum is larger, I move the right pointer left to decrease it. When the sum matches, I return the 1-based indexes. Each pointer moves at most `n` times, so the time complexity is `O(n)` and the extra space is `O(1)`.

## Main takeaways

- Sorted input can turn a pair-search into a linear inward scan.
- The value of two pointers comes from proving that a whole group of candidates can be discarded.
- Prefer one clear pointer decision per loop iteration over nested loops that repeatedly recalculate the same state.
- Return a general-purpose failure value such as `nil` instead of fabricating indexes after the pointers cross.
- Always account for the problem’s index convention: this problem expects 1-based indexes even though Go slices use 0-based indexes.
