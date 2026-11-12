# 11. Container With Most Water — Mock Interview Transcript

> English-speaking practice transcript for [11. Container With Most Water](../11-container-with-most-water.md).
> Both roles are written out so the phrasing can be reused in a real interview.

**Format:** Full transcript (Interviewer 🧑‍💼 / Interviewee 🧑‍💻)
**Focus:** Natural interview English — clarification, approach, code, correctness, complexity, edge cases, follow-ups

---

## Act 1 — Clarification

🧑‍💼 **Interviewer:**
> You’re given an array `height`, where each value is the height of a vertical line at that position. Choose two lines that hold the most water and return the maximum area. What would you clarify before solving it?

🧑‍💻 **Interviewee:**
> I want to confirm the area is determined by the distance between the two positions times the shorter of the two heights. Also, can I assume the input has at least two lines, as in the usual problem constraints? If not, I can return zero when there aren’t two positions.

🧑‍💼 **Interviewer:**
> Yes, that’s the area formula. You may assume the array has at least two values, but handling shorter input safely is fine.

🧑‍💻 **Interviewee:**
> Understood. I’ll compute `min(height[left], height[right]) * (right - left)` and keep the largest area seen.

> 💬 **Why this act matters in English:** Clarifying the area formula gives you a precise objective. *"I want to confirm…"* and *"Can I assume…?"* are useful ways to check the contract without delaying the solution.

---

## Act 2 — Approach

🧑‍💼 **Interviewer:**
> What approach would you use?

🧑‍💻 **Interviewee:**
> A brute-force approach would check every pair of lines, which takes `O(n²)` time. I can do better with inward two pointers.
>
> I’ll start `left` at the first line and `right` at the last. For each pair, I’ll calculate the area using the shorter line as the limiting height and update the best area. Then I’ll move the pointer at the shorter line inward.
>
> The reason is that moving either pointer reduces the width. If I keep the shorter line and move the taller one, the limiting height can’t increase, because the shorter line is still one wall. So the area cannot improve. To have a chance at a larger area, I discard the shorter wall and look for a taller one.

🧑‍💼 **Interviewer:**
> What if both lines have the same height?

🧑‍💻 **Interviewee:**
> Either pointer can move. The current height limit is the same on both sides, and moving one inward is the only way to continue. The code can consistently move `right` on a tie.

🧑‍💼 **Interviewer:**
> How is this different from a sliding window?

🧑‍💻 **Interviewee:**
> This is inward traversal. Both pointers start at opposite ends and move toward each other, and the shorter wall tells me which side to discard. A sliding window usually expands one boundary and shrinks the other to maintain a range property. Here I’m evaluating pairs of boundaries, not maintaining a valid contiguous window.

> 💬 **Phrase bank:** *"The key observation is…"*, *"The shorter line is the limiting factor."*, *"Moving this side cannot improve the result because…"*

---

## Act 3 — Coding

🧑‍💼 **Interviewer:**
> Please implement it in Go and explain the state you’re maintaining.

🧑‍💻 **Interviewee:**
> `left` and `right` define the current candidate container. `bestArea` stores the largest area from every pair I’ve evaluated. I’ll calculate the current area before moving a pointer.

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

> This uses Go’s built-in `min` and `max`, available in Go 1.21 and later. If the project uses an earlier Go version, I can keep the same pointer logic and use an `if area > bestArea` update plus a small minimum comparison.

🧑‍💼 **Interviewer:**
> Why do you update `bestArea` before moving a pointer?

🧑‍💻 **Interviewee:**
> The current pair is still a valid candidate, so I need to calculate its area before I discard one of its boundaries. After moving a pointer, that exact pair is no longer available.

---

## Act 4 — Correctness & Complexity

🧑‍💼 **Interviewer:**
> Convince me that moving the shorter line won’t skip the optimal container.

🧑‍💻 **Interviewee:**
> Suppose the left line is shorter. Any container that keeps this same left line and uses a position to its left is impossible because the right pointer is already the current right boundary of the remaining interval. Any container that keeps the left line and uses a position between the pointers has a smaller width than the current pair. Its height is still limited by that same left line, so its area cannot exceed the current area, which I already recorded.
>
> Therefore, no better container using the shorter left line remains inside the interval. I can discard it and search for a taller left line. The argument is symmetric when the right line is shorter. Each step either records the current area or rules out all remaining candidates that use the shorter boundary.

🧑‍💼 **Interviewer:**
> What are the time and space costs?

🧑‍💻 **Interviewee:**
> Each iteration moves exactly one pointer inward, so there are at most `n - 1` iterations. The time is `O(n)`, and the extra space is `O(1)` because I only keep a few integer variables.

---

## Act 5 — Walkthrough & Edge Cases

🧑‍💼 **Interviewer:**
> Walk through `[1, 8, 6, 2, 5, 4, 8, 3, 7]`.

🧑‍💻 **Interviewee:**
> I start at indexes `0` and `8`. The area is `min(1, 7) * 8 = 8`. The left line is shorter, so I move `left` inward. I keep checking the area and moving the shorter boundary. At indexes `1` and `8`, the area is `min(8, 7) * 7 = 49`, which becomes the best result. Continuing inward cannot remove that recorded answer, so the function eventually returns `49`.

🧑‍💼 **Interviewer:**
> What cases would you check?

🧑‍💻 **Interviewee:**
> I’d check:
>
> | Input | Expected | Why |
> |---|---:|---|
> | `[1, 1]` | `1` | The only possible pair |
> | `[1, 2]` | `1` | Width and shorter height are both one |
> | `[0, 0, 0]` | `0` | Every candidate has zero height |
> | `[5, 1, 4]` | `8` | The best pair may not use the outermost right line after a move |
> | `[]` or one line | `0` | The loop does not run |
>
> I’d also verify that the best area is retained when it occurs before the pointers meet.

> 💬 **Phrase bank:** *"Let me trace the current boundaries."*, *"I record the area before discarding a wall."*, *"The best result remains stored even after the pointers move."*

---

## Act 6 — Follow-ups

🧑‍💼 **Interviewer:**
> Could sorting the heights help?

🧑‍💻 **Interviewee:**
> No. The positions determine the width, so sorting would change which heights are paired with which positions and would invalidate the input geometry. The two-pointer scan uses the original order and is already `O(n)`, which is optimal because it may need to inspect every line.

🧑‍💼 **Interviewer:**
> What if the heights can be negative?

🧑‍💻 **Interviewee:**
> The usual problem defines heights as non-negative. If negative values were allowed, I’d clarify what a negative height means physically and what area should represent. I wouldn’t silently apply the same formula without a meaningful input contract.

🧑‍💼 **Interviewer:**
> What is the main thing you would explain while presenting this solution?

🧑‍💻 **Interviewee:**
> The greedy move needs justification: width always shrinks, so the only useful move is to discard the shorter limiting wall and search for a taller one. That explains both why the algorithm is linear and why it remains correct.

---

## Scorecard (Interviewer's internal notes)

| Dimension | Verdict | Comment |
|---|---|---|
| Clarification | 🟢 Strong | Confirmed the limiting-height area formula and input boundary |
| Pattern recognition | 🟢 Strong | Chose inward traversal rather than pair enumeration |
| Pointer reasoning | 🟢 Strong | Explained why the shorter boundary is safe to discard |
| Correctness | 🟢 Strong | Connected width reduction and limiting height to the discarded candidates |
| Complexity | 🟢 Strong | Counted one pointer move per iteration and stated constant space |
| Edge cases | 🟢 Strong | Included two lines, zero heights, short input, and an interior optimum |
| Communication | 🟢 Strong | Named the state and explained why area is evaluated before movement |

---

## Phrase bank for practice

| Situation | English you can reuse |
|---|---|
| Clarifying | "I want to confirm the area formula." / "Can I assume there are at least two lines?" |
| Proposing | "I’ll start with the widest possible pair and move inward." |
| Explaining the greedy move | "The shorter line limits the water height." / "Keeping it while reducing width cannot improve the area." |
| Correctness | "Any remaining pair using this shorter wall has less width and no greater limiting height." |
| Complexity | "Each iteration moves one boundary, so the scan is linear." |
| Buying time | "Let me compare what happens if I keep this wall and move the other one." |
