# Sliding Windows

## The takeaway

A sliding window is a contiguous range of a sequence, `[left, right]`, that we move across the input instead of restarting the search at every position.

The interview insight is this: **if the answer must be contiguous, and a bigger window can only make the condition worse (or only make it better), we never need to test every start position.** We keep one window, grow it on the right, and shrink it from the left only when it breaks the condition.

That turns a nested double loop, which checks every subarray in $O(n^2)$ or worse, into a single pass that runs in $O(n)$.

Use it when these three things are true:

1. The answer is a **contiguous** subarray or substring.
2. The window condition is **monotone** — growing it moves the condition in one direction, shrinking moves it back.
3. The window state (sum, counts, distinct keys) can be **updated in $O(1)$** when one element enters or leaves.

If any of the three fails, sliding window will produce a wrong answer that still looks plausible. That is why the condition check gets its own section below.

## Before you start

Four terms appear throughout this note:

| Term | Plain meaning |
| --- | --- |
| **Contiguous** | Elements next to each other with no gaps. `[2,3]` is contiguous inside `[1,2,3]`; `[1,3]` is not. |
| **Window** | The current contiguous range `[left, right]`. It is what the two pointers delimit. |
| **Window state** | The small summary you maintain while the window moves: a running sum, a character count, a set of distinct values. |
| **Monotone** | Moves in one direction only. Here: growing a window never makes an invalid condition valid again. |

You do not need any special data structure to start. An array of integers and two `int` indexes are enough. A hash map is only needed when the state is a set of counts, as in "at most `K` distinct characters."

## What is a window?

Take an array and place two markers around a part of it:

```text
nums  = [2, 1, 5, 1, 3, 2]
         ↑        ↑
       left     right

window = [2, 1, 5, 1]   length = right - left + 1 = 4
```

`left` is the first index inside the window. `right` is the last index inside. Everything between them belongs to the window.

A sliding window algorithm moves these markers by rules:

- **Expand:** `right += 1` admits one new element.
- **Shrink:** `left += 1` drops one old element.
- **Both markers only move forward.** They never jump back, which is the reason the algorithm stays linear.

### Why this beats brute force

The brute-force version tries every possible window:

```text
for left from 0 to n - 1:
    for right from left to n - 1:
        check window [left, right]
```

There are about $n^2 / 2$ windows, and recomputing each window's state from scratch can push that to $O(n^3)$.

The sliding version notices something the brute-force version ignores: **two neighboring windows share almost all their elements.**

```text
[2  1  5  1]        →  next window drops 2 and adds 3
   [1  5  1  3]
```

So instead of recounting, we do one subtraction and one addition. That single idea is the whole optimization.

## The two kinds of windows

Almost every sliding window problem is one of these two shapes. Learn both templates and you can start most problems immediately.

| | **Fixed-size window** | **Variable-size window** |
| --- | --- | --- |
| Window length | Always exactly `k` | Changes; found by the algorithm |
| Trigger to shrink | Every step, once the window exceeds `k` | Only when the condition breaks (or is satisfied) |
| Loop inside | `if`, because at most one element leaves | `while`, because many elements may leave |
| Typical question | "Best window of size `k`" | "Longest / shortest window that satisfies a condition" |
| Example | Maximum average subarray | Longest substring without repeats |

### Template A — fixed-size window

Move the window one step at a time by adding the incoming element and removing the outgoing one.

```text
build the window of the first k elements
record the answer

for right from k to n - 1:
    add nums[right]
    remove nums[right - k]      # the element that just fell out of the window
    record the answer
```

```go
func maxSumOfSizeK(nums []int, k int) int {
	if k <= 0 || k > len(nums) {
		return 0
	}

	sum := 0
	for i := 0; i < k; i++ {
		sum += nums[i]
	}
	best := sum

	for right := k; right < len(nums); right++ {
		sum += nums[right]
		sum -= nums[right-k]
		if sum > best {
			best = sum
		}
	}
	return best
}
```

Notice that `right - k` is always the index of the element leaving, so no separate `left` variable is required. Keeping an explicit `left` is also fine and often easier to read.

### Template B — variable-size window

This is the one to memorize. Grow on the right, shrink from the left until the window is acceptable again, then record the answer.

```text
left = 0
state = empty
best = 0

for right from 0 to n - 1:
    add input[right] to state

    while state is not acceptable:
        remove input[left] from state
        left += 1

    best = better(best, right - left + 1)
```

For "longest" problems, the condition is the *good* condition and `while` removes violations.

For "shortest" problems, the loop condition flips: shrink while the window is *still good*, because every good window is a candidate and we want to test whether a smaller one also works.

```text
left = 0
state = empty
best = infinity

for right from 0 to n - 1:
    add input[right] to state

    while state is acceptable:
        best = min(best, right - left + 1)
        remove input[left] from state
        left += 1
```

The two templates differ in only three places: the direction of the comparison, when the answer is recorded, and whether a better answer means longer or shorter. Everything else is identical.

### Which one is this problem?

Ask one question: **does the problem give me the window size?**

- "of size `k`", "of length `k`", "every window of size `k`" → **fixed**.
- "longest", "shortest", "minimum length", "maximum length", "at most `K`" → **variable**.

If the problem gives `k` but asks for the best window, it is still fixed-size. If `k` is a limit on something inside the window (distinct values, flips, replacements), it is variable-size.

## The rule that makes it valid: monotonicity

This is the part most people skip, and the part interviewers test.

Sliding window is correct only when the window condition is **monotone under growth**: once a window is invalid, every larger window that contains it is also invalid.

When that holds, shrinking is the only way back to a valid window, and the `while` loop never skips a valid answer.

### Conditions that are safe

| Problem condition | Why it is monotone |
| --- | --- |
| Sum of a window of **non-negative** numbers ≥ target | Adding more non-negative numbers can only keep the sum at or above the target. |
| At most `K` distinct characters | A superset of a window has at least as many distinct characters. |
| At most `K` zeroes / flips | A superset can only contain more zeroes. |
| No repeated character | A superset with a duplicate keeps the duplicate. |
| Window contains all required characters | A superset still contains all of them. |

### Conditions that are not safe

| Situation | Why sliding window fails | What to use instead |
| --- | --- | --- |
| Maximum subarray sum with **negative** numbers | Adding a negative number can lower the sum, so a "bad" window can become good again. | Kadane's algorithm |
| Longest subarray with sum ≤ target, with negative numbers | Shrinking can make the sum go *up*, so `left` cannot be advanced on a fixed rule. | Prefix sums with a sorted index structure |
| Subarray sum exactly equal to target, with mixed signs | Equality is not monotone in either direction. | Prefix-sum hash map |
| Counting distinct **subsequences** | Subsequences are not contiguous. | Dynamic programming |

If you cannot explain why the condition moves in one direction, do not use a sliding window. Say so in the interview and pick a different approach — that answer scores higher than a broken linear solution.

## Worked example 1 — fixed window

**Problem:** given `nums = [2, 1, 5, 1, 3, 2]` and `k = 3`, find the largest sum of any contiguous subarray of length `k`.

**Window state:** one integer, `sum`.

**Invariant:** before each comparison, `sum` equals the sum of exactly the `k` elements ending at `right`.

```text
[2  1  5] 1  3  2      sum = 8
    [1  5  1] 3  2      sum = 8 - 2 + 1 = 7
          [5  1  3] 2   sum = 7 - 1 + 3 = 9
                [1  3  2]  sum = 9 - 5 + 2 = 6
```

| Step | Window | Added | Removed | `sum` | `best` |
| --- | --- | --- | --- | --- | --- |
| start | `[2,1,5]` | — | — | 8 | 8 |
| 1 | `[1,5,1]` | 1 | 2 | 7 | 8 |
| 2 | `[5,1,3]` | 3 | 1 | 9 | 9 |
| 3 | `[1,3,2]` | 2 | 5 | 6 | 9 |

Answer: `9`.

One detail worth saying out loud in an interview: each element enters the window once and leaves once, so the work is one addition and one subtraction per step. That is why the answer is $O(n)$ even though a naive implementation would rescan `k` elements per window.

## Worked example 2 — variable window

**Problem:** given `nums = [2, 3, 1, 2, 4, 3]` and `target = 7`, find the length of the shortest contiguous subarray whose sum is at least `target`.

This is a "shortest" problem, so we shrink while the window is valid.

```text
left = 0, sum = 0, best = infinity

expand right until sum >= 7
    then drop elements from the left while sum is still >= 7
    every time the window is valid, try to record a shorter length
```

| Step | `right` | Added | `sum` | Shrink actions | `best` |
| --- | --- | --- | --- | --- | --- |
| 1 | 0 | 2 | 2 | none, `2 < 7` | ∞ |
| 2 | 1 | 3 | 5 | none, `5 < 7` | ∞ |
| 3 | 2 | 1 | 6 | none, `6 < 7` | ∞ |
| 4 | 3 | 2 | 8 | length 4, drop 2 → sum 6 | 4 |
| 5 | 4 | 4 | 10 | length 4, drop 3 → 7; length 3, drop 1 → 6 | 3 |
| 6 | 5 | 3 | 9 | length 3, drop 2 → 7; length 2, drop 4 → 3 | **2** |

Answer: `2`, from the subarray `[4, 3]`.

```go
func minSubArrayLen(target int, nums []int) int {
	left, sum := 0, 0
	best := len(nums) + 1 // sentinel meaning "not found yet"

	for right := 0; right < len(nums); right++ {
		sum += nums[right]

		for sum >= target {
			if length := right - left + 1; length < best {
				best = length
			}
			sum -= nums[left]
			left++
		}
	}

	if best > len(nums) {
		return 0 // no window satisfies the condition
	}
	return best
}
```

Two things make this correct:

- The answer is recorded **before** shrinking, because the current window is valid at that moment.
- The inner `while` keeps shrinking even after the first valid window, because a shorter window may still be valid.

### Longest and shortest are not symmetric

| | Longest valid window | Shortest valid window |
| --- | --- | --- |
| Inner loop runs | while the window is **invalid** | while the window is **valid** |
| Record the answer | after shrinking, window is valid | before shrinking, window is valid |
| Can `if` replace `while`? | Yes, see below | No, always `while` |
| Initial `best` | `0` | a sentinel larger than any real length |

The `if` shortcut for longest problems is a common optimization: the best length can grow by at most one per step, so shrinking once is enough to keep the window from falling below the current best. It is correct but harder to reason about, so write `while` first and only switch to `if` if you are confident.

## Why the total time is $O(n)$

The inner `while` loop looks like it could make the algorithm quadratic. It does not, because of a simple counting argument:

- `right` moves from `0` to `n - 1`. Total moves: `n`.
- `left` only ever moves **forward**, from `0` to at most `n - 1`. Total moves: at most `n`.
- Every iteration of the inner `while` moves `left` exactly once.

So the inner loop body can run at most `n` times across the entire algorithm, not `n` times per outer step. Total pointer moves ≤ `2n`, which is $O(n)$.

This is called **amortized** analysis: any single step may do a lot of shrinking, but the total is bounded.

| Approach | Time | Extra space |
| --- | --- | --- |
| Check every subarray, recount each | $O(n^3)$ | $O(1)$ |
| Check every subarray, incremental state | $O(n^2)$ | $O(1)$ |
| Sliding window | $O(n)$ | $O(1)$ for sums, $O(k)$ for count maps |

The space cost is only for the window state. A running sum or a single counter needs constant space. A map of character counts needs space proportional to the alphabet or to `k`.

## When to look for a sliding window

Strong signals:

- The words **contiguous**, **subarray**, **substring**, or **window** appear.
- The question asks for a **longest**, **shortest**, **maximum**, **minimum**, or **count of** ranges.
- A size or limit `k` is given: window size, allowed replacements, allowed flips, distinct values.
- The input is a sequence: array, string, or stream.
- The state can be maintained incrementally as elements enter and leave.

Weak or false signals — these look similar but usually are not sliding windows:

- "Any subsequence" — not contiguous.
- "Count subarrays with sum exactly `K`" and the array has negative numbers — use prefix sums.
- "Maximum subarray sum" with no size limit — use Kadane's algorithm.
- Pair problems on a **sorted** array — that is two pointers, not a window.

The two-pointer and sliding-window families overlap: a window *is* two pointers. The difference is intent. Two pointers compare or coordinate two positions; a window maintains a valid contiguous region and asks how it can grow or shrink.

## Quick pattern-selection guide

| If the problem says... | Start by asking... |
| --- | --- |
| "every window of size `k`" | Can I update the state by adding one element and removing one? |
| "longest substring with at most `K` ..." | Which counter breaks the limit, and which element do I drop to fix it? |
| "shortest subarray with sum ≥ target" | Are all values non-negative? If yes, shrink while valid. |
| "exactly `K` ..." | Can I compute "at most `K`" minus "at most `K - 1`"? |
| "maximum in every window of size `k`" | This name is misleading — use a monotonic deque, not a window state. |
| "minimum window containing all characters" | Shrink while valid, and track how many required characters are still missing. |

### The "exactly `K`" trick

Conditions with *exactly* are not monotone, so the templates do not apply directly. Use subtraction instead:

```text
count(exactly K) = count(at most K) - count(at most K - 1)
```

Each term is a monotone "at most" problem, so each one is a normal sliding window. This pattern appears in problems such as counting subarrays with exactly `K` distinct integers.

## Interview-solving process

Use this sequence when explaining a sliding window solution:

1. **Confirm the answer is contiguous.** If it is not, stop this pattern.
2. **Confirm monotonicity.** Say which direction the condition moves and why growing cannot restore a broken window.
3. **Name the window state.** "A sum," "a count of distinct characters," "a map from character to frequency." Say what each part means.
4. **State the invariant.** For example: "`[left, right]` is the longest valid window ending at `right`."
5. **Choose fixed or variable, and `if` or `while`.** Justify the inner loop's condition.
6. **Say where the answer is recorded.** Before shrinking for shortest, after shrinking for longest.
7. **Analyze complexity and edge cases.** Each element enters and leaves once, so $O(n)$. Check empty input, `k` larger than the input, `k <= 0`, one-element input, no valid window, all elements identical, and possible integer overflow on large sums.

A concise spoken answer:

> The answer is contiguous, and the condition is monotone: adding characters can only add distinct values, never remove them. So I keep a window `[left, right]` with a count map. For each `right`, I add the character, then shrink `left` while the window has more than `K` distinct values. After shrinking, the window is the longest valid one ending at `right`, so I update the answer. Each index enters and leaves the window once, so this is `O(n)` time and `O(K)` space.

## Common mistakes

| Mistake | Symptom | Fix |
| --- | --- | --- |
| Forgetting to remove the outgoing element | Window length grows past `k`; sum is too large | In a fixed window, always add then subtract `nums[right-k]` |
| Using `if` where `while` is needed | Shortest-window answers are too large | Shortest problems must loop while valid |
| Recording the answer at the wrong time | Off-by-one in length | Shortest: record before shrinking. Longest: record after |
| Using `right - left` as length | Every answer is one too small | Length is `right - left + 1` |
| Not checking `prev >= left` for repeat detection | An old occurrence outside the window forces `left` forward incorrectly | Compare stored indexes against the current `left` |
| Sliding over a non-monotone condition | Passes the samples, fails on negatives or equality | Verify monotonicity before coding |
| Ignoring stale map entries | Distinct count never decreases | Decrement counts on removal and delete keys at zero |
| Assuming `s[i]` is one character in Go | Breaks on non-ASCII input | `s[i]` is a byte; convert to `[]rune` for Unicode |
| Initializing `best` to `0` for shortest | Zero is returned when no window exists | Use a sentinel and check it at the end |

## Practice ladder

Work through these in order. Each one adds exactly one new idea.

| # | Problem | Window | New idea |
| --- | --- | --- | --- |
| 1 | [643. Maximum Average Subarray I](https://leetcode.com/problems/maximum-average-subarray-i/) | Fixed | Build the sum once, then slide |
| 2 | [1456. Maximum Number of Vowels in a Substring of Given Length](https://leetcode.com/problems/maximum-number-of-vowels-in-a-substring-of-given-length/) | Fixed | Maintain a count instead of a sum |
| 3 | [209. Minimum Size Subarray Sum](./209-minimum-size-subarray-sum.md) ([LeetCode](https://leetcode.com/problems/minimum-size-subarray-sum/)) | Variable, shortest | Shrink while valid |
| 4 | [3. Longest Substring Without Repeating Characters](https://leetcode.com/problems/longest-substring-without-repeating-characters/) | Variable, longest | Track the last index of each character |
| 5 | [567. Permutation in String](https://leetcode.com/problems/permutation-in-string/) | Fixed | Compare two frequency maps |
| 6 | [1004. Max Consecutive Ones III](https://leetcode.com/problems/max-consecutive-ones-iii/) | Variable, longest | The `if`-instead-of-`while` optimization |
| 7 | [424. Longest Repeating Character Replacement](https://leetcode.com/problems/longest-repeating-character-replacement/) | Variable, longest | A condition built from the maximum count |
| 8 | [438. Find All Anagrams in a String](https://leetcode.com/problems/find-all-anagrams-in-a-string/) | Fixed | Record every valid window, not just the best |
| 9 | [76. Minimum Window Substring](https://leetcode.com/problems/minimum-window-substring/) | Variable, shortest | Two maps and a "missing count" |
| 10 | [992. Subarrays with K Different Integers](https://leetcode.com/problems/subarrays-with-k-different-integers/) | Variable, counting | `atMost(K) - atMost(K - 1)` |
| 11 | [239. Sliding Window Maximum](https://leetcode.com/problems/sliding-window-maximum/) | Fixed | Not a window state — a monotonic deque |

For each problem, write down three lines before coding: the window state, the invariant, and the shrink trigger. If you cannot fill in all three, the pattern does not fit yet.

## Edge cases to test

- Empty input.
- One element, both when it satisfies the condition and when it does not.
- `k` equal to `1`, `k` equal to `n`, and `k` greater than `n`.
- All elements identical.
- No valid window exists at all.
- A valid window that starts at index `0` or ends at the last index.
- Sum or product values large enough to overflow a 32-bit integer.

## Related patterns

- **Two pointers:** The parent technique. A window is two pointers plus a validity rule. See [Two Pointers](../two-pointers/).
- **Fast and slow pointers:** Two pointers that move at different speeds to find cycles or midpoints. No contiguous state is involved.
- **Prefix sums:** The right tool when the window condition is not monotone, such as subarray sums with negative numbers. Combine with a hash map for "exactly `K`" and with a deque for range minima and maxima.
- **Monotonic deque:** Keeps the maximum or minimum of a sliding window in $O(1)$ amortized time, replacing the "rescan the window" step.

## Source and further reading

This note is an original study guide informed by [ByteByteGo's Introduction to Sliding Windows](https://bytebytego.com/courses/coding-patterns/sliding-windows/introduction-to-sliding-windows). The chapter presents sliding windows as a specialized extension of two pointers, covering fixed-size windows, dynamic windows, and the conditions under which each variant stays linear.

## Next steps

1. Re-derive both templates from memory on a blank page. If you can write the expand-shrink skeleton without notes, the pattern is learned.
2. Solve problems 1–4 from the practice ladder and time yourself. State the invariant out loud before writing code.
3. For each solution, answer one extra question: why is this $O(n)$ and not $O(n^2)$?
4. Attempt problem 10 or 11 only after problems 1–9 feel mechanical. Both break the simple template in a specific way, and that is the point.
