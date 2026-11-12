# 15. 3Sum — Mock Interview Transcript

> English-speaking practice transcript for [15. 3Sum](../15-3sum.md).
> Both roles are written out so the phrasing can be reused in a real interview.

**Format:** Full transcript (Interviewer 🧑‍💼 / Interviewee 🧑‍💻)
**Focus:** Natural interview English — clarification, approach, code, correctness, complexity, edge cases, follow-ups

---

## Act 1 — Clarification

🧑‍💼 **Interviewer:**
> Given an integer array `nums`, return all unique triplets `[nums[i], nums[j], nums[k]]` such that the three values sum to zero. Before you start, what would you like to clarify?

🧑‍💻 **Interviewee:**
> I want to confirm that the three values must come from three distinct indexes, even if some values in the array are equal. And when you say unique triplets, do you mean that two triplets with the same values count as duplicates regardless of their indexes?
>
> Also, can I return the triplets in any order? And should I preserve the input array, or may I sort it in place?

🧑‍💼 **Interviewer:**
> Correct: distinct indexes, unique value combinations, and any output order is fine. You may sort the input in place.

🧑‍💻 **Interviewee:**
> Great. So I need to return each value combination once, with three distinct elements, and I can modify the input by sorting it.

> 💬 **Why this act matters in English:** phrases like *"I want to confirm…"*, *"Do you mean…?"*, and *"May I…?"* clarify the contract before you commit to an approach.

---

## Act 2 — Approach

🧑‍💼 **Interviewer:**
> Talk me through your first idea and how you would improve it.

🧑‍💻 **Interviewee:**
> The direct approach is to examine every combination of three indexes and check whether its values sum to zero. That takes `O(n³)` time. I would also need a set or another mechanism to remove duplicate value combinations.
>
> Since I’m allowed to sort, I can do better. I’ll sort the array, fix one value as an anchor, then search the suffix to its right with two pointers for a pair that sums to the negative of the anchor. This gives an outer loop over anchors and a linear scan for each anchor, so the search is `O(n²)`. Sorting is `O(n log n)`, which is dominated by the search.
>
> For each anchor, `left` starts at `i + 1` and `right` starts at the last index. If the sum is too small, I move `left` right to get a larger value. If it’s too large, I move `right` left to get a smaller value. If it’s zero, I save the triplet, move both pointers, and skip duplicates.

🧑‍💼 **Interviewer:**
> Why is it safe to move `left` when the sum is too small?

🧑‍💻 **Interviewee:**
> The array is sorted in ascending order. If the current sum is too small, keeping `left` and moving `right` left would use an equal or smaller value, so the sum could only stay the same or decrease. That cannot fix the problem. I need a larger value at `left`, so I move it right.
>
> Conversely, if the sum is too large, keeping `right` and moving `left` right would use an equal or larger value, so that can’t reduce the sum. I move `right` left to try a smaller value.

🧑‍💼 **Interviewer:**
> What type of two-pointer strategy is this? Is it a sliding window?

🧑‍💻 **Interviewee:**
> It’s **inward traversal** on the sorted suffix. The pointers start at opposite ends of the active range and move toward each other. It isn’t a sliding window because I’m not maintaining a contiguous range that expands and shrinks to satisfy a window condition. The ordering lets each pointer move rule out candidates safely.

> 💬 **Phrase bank:** *"The straightforward approach would be…"*, *"The key observation is…"*, *"This is safe because…"*, *"I’d distinguish it from…"*

---

## Act 3 — Coding

🧑‍💼 **Interviewer:**
> Let’s see it in Go. Please explain how you’ll handle duplicates as you write.

🧑‍💻 **Interviewee:**
> I’ll skip an anchor if it has the same value as the previous anchor. When I find a match, I’ll append it, move both pointers first, and then skip any additional copies of the values I just used. Moving first makes sure that the match branch always makes progress.

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	result := make([][]int, 0)

	for i := 0; i < len(nums)-2; i++ {
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}

		left, right := i+1, len(nums)-1
		for left < right {
			sum := nums[i] + nums[left] + nums[right]

			switch {
			case sum < 0:
				left++
			case sum > 0:
				right--
			default:
				result = append(result, []int{nums[i], nums[left], nums[right]})
				left++
				right--

				for left < right && nums[left] == nums[left-1] {
					left++
				}
				for left < right && nums[right] == nums[right+1] {
					right--
				}
			}
		}
	}

	return result
}
```

> This needs `import "sort"`. Since `sort.Ints` changes the input slice, I’d copy `nums` first if the caller needs its original order preserved.

🧑‍💼 **Interviewer:**
> Why use `continue` for a duplicate anchor instead of incrementing `i` manually?

🧑‍💻 **Interviewee:**
> The `for` loop already increments `i`. If I increment it manually and then continue, the loop step increments it again, which could skip an anchor value and miss a valid triplet. `continue` lets the normal loop step advance exactly once.

🧑‍💼 **Interviewer:**
> Why do you move both pointers before skipping duplicates after a match?

🧑‍💻 **Interviewee:**
> The matched pair itself must not be examined again. Advancing both pointers guarantees we leave that pair even if the next values are different. Then the loops skip further copies of the values used in the triplet. If I only tried to skip duplicates without first moving, I could stay on the same pair and append it repeatedly.

---

## Act 4 — Correctness & Complexity

🧑‍💼 **Interviewer:**
> Convince me that the scan won’t miss a valid pair for an anchor.

🧑‍💻 **Interviewee:**
> For a fixed anchor, the active candidate pairs are inside `[left, right]`. If the sum is too small, the current left value cannot work with any value at or left of `right`, because those values are no larger than the current right value. So I can discard that left position. If the sum is too large, the current right value cannot work with any value at or right of `left`, because those values are no smaller than the current left value. So I can discard that right position.
>
> A zero sum is recorded. I then advance both pointers to continue with a different pair. Each step either finds a result and advances, or rules out a pointer position, so the scan terminates. Skipping equal anchors or pair values removes only combinations with values already considered; it doesn’t remove a distinct triplet.

🧑‍💼 **Interviewer:**
> What’s the complexity?

🧑‍💻 **Interviewee:**
> Sorting takes `O(n log n)`. There are up to `n` anchors, and for each one the two pointers move across at most `n` positions, so the search takes `O(n²)`. Total time is `O(n²)`, with sorting dominated by the search.
>
> The scan uses `O(1)` auxiliary space, excluding sorting implementation details. The returned triplets take additional output space depending on how many unique results there are.

---

## Act 5 — Walkthrough & Edge Cases

🧑‍💼 **Interviewer:**
> Walk through `[-1, 0, 1, 2, -1, -4]`.

🧑‍💻 **Interviewee:**
> After sorting, the values are `[-4, -1, -1, 0, 1, 2]`.
>
> With anchor `-4`, the remaining pairs are all too large to bring the total to zero, so that scan finds no result. With the first `-1`, the scan finds `[-1, -1, 2]` and `[-1, 0, 1]`. The next `-1` anchor is skipped because it duplicates the previous anchor. The later anchors produce no additional triplets. So the result contains those two unique combinations.

🧑‍💼 **Interviewer:**
> What edge cases would you test?

🧑‍💻 **Interviewee:**
> I’d check:
>
> | Input | Expected | Why |
> |---|---|---|
> | `[]` or fewer than three values | `[]` | There aren’t enough elements for a triplet |
> | `[1, 2, 3]` | `[]` | No zero-sum combination |
> | `[0, 0, 0]` | `[[0, 0, 0]]` | All values match; the triplet should appear once |
> | `[0, 0, 0, 0]` | `[[0, 0, 0]]` | Repeated values must not duplicate the result |
> | `[-2, 0, 1, 1, 2]` | `[[-2, 0, 2], [-2, 1, 1]]` | The same anchor can produce different pairs |
>
> The all-zero input is especially useful because it checks both anchor deduplication and pair deduplication. It also catches a match branch that fails to move its pointers.

> 💬 **Phrase bank:** *"Let me trace that after sorting."*, *"The case I’d be most careful about is…"*, *"That checks both…"*

---

## Act 6 — Follow-ups

🧑‍💼 **Interviewer:**
> Suppose the target were a value `target` instead of zero. What changes?

🧑‍💻 **Interviewee:**
> The structure stays the same. I’d compare `nums[i] + nums[left] + nums[right]` with `target`. A sum below the target moves `left` right, and a sum above it moves `right` left. The anchor and pair duplicate rules still apply, assuming the output should contain unique value triplets.

🧑‍💼 **Interviewer:**
> Could you keep the original order of `nums`?

🧑‍💻 **Interviewee:**
> Yes. I’d make a copy, sort the copy, and run the same algorithm on it. That adds `O(n)` space for the copy. The returned result and the `O(n²)` time bound stay the same.

🧑‍💼 **Interviewer:**
> What if the input is extremely large and adding three integers could overflow?

🧑‍💻 **Interviewee:**
> I’d check the input constraints against Go’s `int` range. If the sum could overflow, I’d convert the values to a wider type before adding them. The sorting and pointer logic would stay the same.

---

## Scorecard (Interviewer's internal notes)

| Dimension | Verdict | Comment |
|---|---|---|
| Clarification | 🟢 Strong | Confirmed distinct indexes, uniqueness, output order, and whether in-place sorting is allowed |
| Pattern recognition | 🟢 Strong | Improved the cubic search by fixing an anchor and scanning with inward pointers |
| Pointer reasoning | 🟢 Strong | Connected each move to sorted order and ruled-out candidates |
| Deduplication | 🟢 Strong | Explained anchor skipping, pair skipping, and why movement must happen first |
| Correctness | 🟢 Strong | Maintained the active interval and accounted for every discarded position |
| Complexity | 🟢 Strong | Included sorting, nested search costs, auxiliary space, and output space |
| Communication | 🟢 Strong | Gave a concrete sorted trace and targeted the all-zero boundary case |

---

## Phrase bank for practice

| Situation | English you can reuse |
|---|---|
| Clarifying | "I want to confirm the contract before I start." / "Do duplicate values at different indexes count as the same triplet?" |
| Proposing | "The straightforward approach would be…" / "I can improve that by sorting and fixing one value at a time." |
| Defining state | "For this anchor, `left` and `right` bound the pairs I haven’t ruled out yet." |
| Justifying a move | "Keeping the other pointer fixed, moving this one in that direction is the only move that can bring the sum closer to the target." |
| Deduplicating | "I’ll advance past the matched pair first, then skip any equal values." |
| Complexity | "There are n anchors, and each scan is linear, so the search is quadratic." |
| Buying time | "Let me trace that on a small sorted example." / "Let me check the all-equal boundary case." |
