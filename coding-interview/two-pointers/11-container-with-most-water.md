# 11. Container With Most Water

## Quick Review

Use two pointers at the ends of `height`. The area between them is:

```text
min(height[left], height[right]) * (right - left)
```

After checking that area, move the pointer at the **shorter line** inward. The shorter line limits the current water level; moving the taller line cannot raise that limit, and it also makes the width smaller.

The accepted two-pointer algorithm is already optimal for this problem:

- **Time:** `O(n)` — each pointer moves inward at most `n` times.
- **Extra space:** `O(1)`.
- **Main improvement opportunity:** make the variable names and pointer-movement explanation clearer. The algorithm itself does not need a different approach.

## Problem

Given an array `height`, where `height[i]` is the height of a vertical line at position `i`, choose two lines that hold the most water. Return the maximum area.

The container's width is the distance between the two positions. Its water height is limited by the shorter line, so:

```text
area = shorter line height * distance between lines
```

Example:

```text
height = [1, 8, 6, 2, 5, 4, 8, 3, 7]
maximum area = 49
```

The maximum comes from the lines at indexes `1` and `8` (0-based): `min(8, 7) * (8 - 1) = 49`.

## Pattern recognition

This is an inward two-pointer problem:

- The two lines define the current container.
- The width is determined by the distance between the pointers.
- The shorter line determines the maximum possible water height.
- Moving a pointer inward reduces width, so a move is useful only if it might find a taller limiting line.

## Your accepted solution

Your solution applies the right algorithm:

```go
func maxArea(height []int) int {
	left, right := 0, len(height)-1

	var res int
	for left < right {
		cur := min(height[left], height[right]) * (right - left)
		if cur > res {
			res = cur
		}

		if height[left] < height[right] {
			left++
		} else {
			right--
		}
	}
	return res
}
```

It checks the current area, keeps the best result, and moves the shorter side. When the heights are equal, moving either pointer is correct; this code moves `right`.

## Can it be better?

### Algorithm and performance

The algorithm is asymptotically optimal. In the worst case, any solution must inspect the input lines, so `O(n)` time is the expected target. The current algorithm also uses only a fixed number of variables, giving `O(1)` extra space.

There is no need to sort the heights: sorting would change the positions that determine width and would add work.

### Readability

The main style opportunity is to use names that describe what each value means:

- `res` → `maxArea` or `bestArea`
- `cur` → `area`

You can also write the result update as `maxArea = max(maxArea, area)` if the Go version provides the built-in `min` and `max` functions. The current `if` statement is equally clear and works without depending on those built-ins.

## Why move the shorter line?

Suppose `height[left] <= height[right]`. The current area is limited by `height[left]`:

```text
current area = height[left] * (right - left)
```

If we keep `left` and move `right` inward, the width gets smaller. The limiting height cannot become greater than `height[left]`, because `height[left]` is still one of the two walls. Therefore, this move cannot produce a larger area than the current one.

To have a chance to improve the area, discard the shorter wall and look for a taller one. The same reasoning applies when the right wall is shorter. If both heights are equal, either side can be moved.

## Recommended version

This version keeps your algorithm and makes the state names more descriptive:

```go
func maxArea(height []int) int {
	left, right := 0, len(height)-1
	bestArea := 0

	for left < right {
		width := right - left
		limitingHeight := min(height[left], height[right])
		area := limitingHeight * width
		bestArea = max(bestArea, area)

		if height[left] < height[right] {
			left++
		} else {
			right--
		}
	}

	return bestArea
}
```

The `min` and `max` built-ins are available in Go 1.21 and later. If the project uses an earlier Go version, keep the explicit result-update `if` or define small helpers. The pointer logic is unchanged.

## Correctness invariant

At the start of each iteration:

> `bestArea` is the largest area found among all containers checked so far, and the pointers mark the current candidate interval.

The current interval is evaluated before a pointer moves. Then the shorter wall is discarded because pairing it with any line inside the interval cannot beat the area already checked for that short wall: the width would be smaller, and the limiting height cannot exceed the short wall. Thus, no potentially better container is lost. When the pointers meet, every candidate that could improve the result has been considered or safely discarded.

## Complexity

- **Time:** `O(n)`.
- **Extra space:** `O(1)`.

Each iteration moves exactly one pointer inward, so there are at most `n - 1` iterations.

## Edge cases

| Case | Result or behavior |
| --- | --- |
| Two lines | Their area is the only possible container. |
| Empty or one-line input | The loop does not run, so the function returns `0`. The original problem normally guarantees at least two lines. |
| Equal heights at both pointers | Move either pointer; the accepted code moves `right`. |
| Zero-height line | Any container using that line has area `0`; moving the shorter side continues the search. |
| Maximum area occurs before the final pair | `bestArea` preserves it even as the pointers continue moving. |

## Interview answer template

> I’ll start with one pointer at each end and compute the area from the width and the shorter line. I move the shorter line inward because keeping it while reducing the width cannot improve the area; moving the taller line has no chance to raise the limiting height. I track the largest area seen. Each step moves one pointer, so the time is `O(n)` and the extra space is `O(1)`.

## Main takeaways

- The shorter line limits the water height.
- Each pointer move reduces width, so move the side that may allow a greater height.
- Explain why the discarded side cannot lead to a better answer; this is the key to justifying the greedy pointer movement.
- Your accepted algorithm already has optimal time and space complexity. The useful refinement is clearer naming and a concise proof.
