# 167. Two Sum II — Input Array Is Sorted — Mock Interview Transcript

> English-speaking practice transcript for [167. Two Sum II — Input Array Is Sorted](../167-two-sum2-input-array-is-sorted.md).
> Both roles are written out so the phrasing can be reused in a real interview.

**Format:** Full transcript (Interviewer 🧑‍💼 / Interviewee 🧑‍💻)
**Focus:** Natural interview English — clarification, approach, code, correctness, complexity, edge cases, follow-ups

---

## Act 1 — Clarification

🧑‍💼 **Interviewer:**
> You’re given a 1-indexed sorted array of integers and a target. Return the two indexes whose values add up to the target. What would you clarify?

🧑‍💻 **Interviewee:**
> I want to confirm that the array is sorted in non-decreasing order and that there is exactly one valid pair, as stated in the standard problem. I also want to confirm the returned indexes must be 1-based, even though Go slices use 0-based indexes. And we can’t use the same element twice, correct?

🧑‍💼 **Interviewer:**
> Correct on all three points.

🧑‍💻 **Interviewee:**
> Great. I’ll use the sorted order to find the pair and add one when I return each slice index.

> 💬 **Why this act matters in English:** State the contract you’ll rely on. In this problem, sorted order enables the optimization and the 1-based requirement is an easy implementation detail to miss.

---

## Act 2 — Approach

🧑‍💼 **Interviewer:**
> How would you solve it?

🧑‍💻 **Interviewee:**
> A straightforward approach checks every pair, which takes `O(n²)` time. Because the array is already sorted, I can use an inward two-pointer scan in `O(n)` time.
>
> I’ll place `left` at the first element and `right` at the last. If their sum is below the target, I move `left` right, because later values are larger and can increase the sum. If the sum is above the target, I move `right` left, because earlier values are smaller and can decrease the sum. When the sum matches, I return the two positions after converting them to 1-based indexes.

🧑‍💼 **Interviewer:**
> Why not sort the array first?

🧑‍💻 **Interviewee:**
> It’s already sorted, so sorting would add unnecessary `O(n log n)` work. It could also change the original positions, which matter because the answer asks for indexes. I can solve it directly in one scan.

🧑‍💼 **Interviewer:**
> What kind of two-pointer traversal is this?

🧑‍💻 **Interviewee:**
> It’s inward traversal. The pointers start at opposite ends, and sorted order tells me which boundary to move. It isn’t a sliding window because I’m not expanding a range to maintain a window condition.

> 💬 **Phrase bank:** *"Since the input is already sorted…"*, *"That move can only increase the sum."*, *"I don’t need to reorder the input."*

---

## Act 3 — Coding

🧑‍💼 **Interviewer:**
> Please implement it in Go.

🧑‍💻 **Interviewee:**
> I’ll compute the sum once per iteration and make exactly one pointer move when it doesn’t match. The problem guarantees an answer, but I’ll return `nil` after the loop as a safe fallback if that guarantee doesn’t hold.

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

🧑‍💼 **Interviewer:**
> Why does the function return `left + 1` and `right + 1`?

🧑‍💻 **Interviewee:**
> Go slice indexes start at zero, but the problem asks for 1-based positions. Adding one converts each index into the required position.

🧑‍💼 **Interviewer:**
> Could the two pointers ever refer to the same element?

🧑‍💻 **Interviewee:**
> The loop condition is `left < right`, so they must refer to distinct indexes whenever I calculate the sum. Once they meet, the loop stops.

---

## Act 4 — Correctness & Complexity

🧑‍💼 **Interviewer:**
> Explain why moving one pointer doesn’t skip the answer.

🧑‍💻 **Interviewee:**
> I maintain the invariant that if a valid pair still exists, its indexes are inside the current inclusive range `[left, right]`.
>
> If the current sum is too small, every pair using `numbers[left]` with an index up to `right` is also too small: those other values are no larger than `numbers[right]`. So the current left position cannot be part of a solution, and I can discard it. If the sum is too large, every pair using `numbers[right]` with an index at least `left` is also too large, because those values are no smaller than `numbers[left]`. So I can discard the right position. When the sum matches, I return the pair. Each move removes only a position that cannot belong to a valid pair in the remaining range.

🧑‍💼 **Interviewer:**
> What are the complexity bounds?

🧑‍💻 **Interviewee:**
> Each pointer only moves inward, and each iteration moves one of them. There are at most `n - 1` iterations, so time is `O(n)`. The scan uses `O(1)` extra space, apart from the returned two-element slice.

---

## Act 5 — Walkthrough & Edge Cases

🧑‍💼 **Interviewer:**
> Walk through `[2, 7, 11, 15]` with target `9`.

🧑‍💻 **Interviewee:**
> I start with values `2` and `15`. Their sum is `17`, which is too large, so I move `right` left. Now the values are `2` and `11`, with sum `13`, still too large, so I move `right` left again. The values are now `2` and `7`, which sum to `9`. Their zero-based indexes are `0` and `1`, so I return `[1, 2]`.

🧑‍💼 **Interviewer:**
> What would you test?

🧑‍💻 **Interviewee:**
> I’d check:
>
> | Numbers | Target | Expected | Why |
> |---|---:|---|---|
> | `[2, 7]` | `9` | `[1, 2]` | Smallest valid input |
> | `[-3, 0, 4, 8]` | `1` | Negative values work with sorted order |
> | `[1, 1, 3, 5]` | `2` | Equal values at distinct indexes may form the answer |
> | `[1, 2, 3, 4]` | `7` | The answer may be found after moving the left pointer |
> | `[2, 7, 11, 15]` | `100` | No answer; the standard problem excludes this case |
>
> I’d pay particular attention to index conversion and to not requiring the two values themselves to be different when they come from different indexes.

> 💬 **Phrase bank:** *"The sum is too large, so I’ll move the right boundary."*, *"These are zero-based indexes, so I’ll add one for the result."*

---

## Act 6 — Follow-ups

🧑‍💼 **Interviewer:**
> What if the input were not sorted?

🧑‍💻 **Interviewee:**
> This pointer-movement proof would no longer apply because moving an index wouldn’t predict whether the value gets larger or smaller. I could use a hash map to find a complement in expected `O(n)` time and keep original indexes, or sort value-index pairs and use two pointers if I need to retain positions. The best choice depends on the space and ordering requirements.

🧑‍💼 **Interviewer:**
> What if no pair exists and there is no guarantee of a solution?

🧑‍💻 **Interviewee:**
> The loop eventually ends when the pointers meet, and I return `nil`. If the API needs to distinguish “not found” from other outcomes, I’d choose an explicit result shape such as a pair plus a boolean, depending on the surrounding codebase.

🧑‍💼 **Interviewer:**
> Does the algorithm still work with negative numbers?

🧑‍💻 **Interviewee:**
> Yes. The reasoning depends on the array’s sorted order, not on whether the values are positive. Moving `left` right never decreases the value, and moving `right` left never increases it, so the sum comparisons still guide the search.

---

## Scorecard (Interviewer's internal notes)

| Dimension | Verdict | Comment |
|---|---|---|
| Clarification | 🟢 Strong | Confirmed sorted input, unique solution guarantee, distinct indexes, and 1-based output |
| Pattern recognition | 🟢 Strong | Used the sorted order for an inward linear scan |
| Pointer reasoning | 🟢 Strong | Explained which positions are ruled out for a low or high sum |
| Implementation | 🟢 Strong | Computes the sum once, uses `left < right`, and converts indexes correctly |
| Complexity | 🟢 Strong | Stated linear time and constant auxiliary space |
| Edge cases | 🟢 Strong | Covered duplicates, negatives, two elements, and the no-solution contract |
| Communication | 🟢 Strong | Kept the explanation tied to the ordering property |

---

## Phrase bank for practice

| Situation | English you can reuse |
|---|---|
| Clarifying | "Can I rely on the array being sorted?" / "Are the returned positions 1-based?" |
| Proposing | "I’ll start with the smallest and largest values." |
| Explaining movement | "Because the array is sorted, moving this pointer in that direction can only…" |
| Correctness | "If a valid pair remains, it must be inside the current range." |
| Index convention | "The slice index is zero-based, so I’ll add one for the requested result." |
| Complexity | "Both pointers move only inward, so the total scan is linear." |
| Buying time | "Let me check which positions this sum comparison rules out." |
