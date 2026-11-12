# 392. Is Subsequence — Mock Interview Transcript

> English-speaking practice transcript for [392. Is Subsequence](../392-is-subsequence.md).
> Both roles are written out so the phrasing can be reused in a real interview.

**Format:** Full transcript (Interviewer 🧑‍💼 / Interviewee 🧑‍💻)
**Focus:** Natural interview English — clarification, approach, code, correctness, complexity, edge cases, follow-ups

---

## Act 1 — Clarification

🧑‍💼 **Interviewer:**
> Given strings `s` and `t`, determine whether `s` is a subsequence of `t`. What would you clarify?

🧑‍💻 **Interviewee:**
> I want to make sure we mean that the characters of `s` must appear in `t` in the same order, but they don’t need to be adjacent. So this is different from checking whether `s` is a substring. Should I assume the strings use the lowercase English letters from the usual problem constraints?

🧑‍💼 **Interviewer:**
> Correct on the subsequence definition. You may assume lowercase English letters.

🧑‍💻 **Interviewee:**
> Great. I’ll scan `t` in order and keep track of the next character I need to match from `s`. If `s` is empty, it’s automatically a subsequence.

> 💬 **Why this act matters in English:** A subsequence preserves order but permits gaps. Saying that distinction clearly prevents you from solving the more restrictive substring problem by mistake.

---

## Act 2 — Approach

🧑‍💼 **Interviewer:**
> What is your approach?

🧑‍💻 **Interviewee:**
> I’ll use a one-way two-pointer scan. `j` scans through `t`, while `i` points to the next character in `s` that I still need to match. If `t[j]` doesn’t match `s[i]`, I advance only `j` and keep looking. If they match, I advance `i`; then I continue scanning `t` for the next required character.
>
> When `i` reaches `len(s)`, I’ve matched every character in order and can return true. If I finish `t` first, then some character from `s` was never found and I return false. The scan is `O(len(t))` time and `O(1)` extra space.

🧑‍💼 **Interviewer:**
> Why is it safe to skip a character in `t` when it doesn’t match?

🧑‍💻 **Interviewee:**
> The current required character is `s[i]`. This `t[j]` is not equal to it, and I cannot come back to this position later without violating the order. It cannot match a later character in `s` first, because subsequence characters must be matched in order. So I skip it and continue forward.

🧑‍💼 **Interviewer:**
> Which two-pointer pattern is this?

🧑‍💻 **Interviewee:**
> It’s unidirectional traversal. Both indexes move forward, but they have different jobs: `j` explores the larger string and `i` records how much of the target string has matched. It isn’t inward traversal or a sliding window.

> 💬 **Phrase bank:** *"I’ll track the next character I need."*, *"This character can be skipped without changing the order."*, *"The two pointers move forward, but serve different roles."*

---

## Act 3 — Coding

🧑‍💼 **Interviewer:**
> Please implement it in Go.

🧑‍💻 **Interviewee:**
> I’ll let `j` scan `t` and stop early if `i` reaches the end of `s`. Then the final check directly says whether all characters were matched. This also handles an empty `s` without a special branch.

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

🧑‍💼 **Interviewer:**
> Why don’t you increment `i` when the characters differ?

🧑‍💻 **Interviewee:**
> `i` represents the next character I still need from `s`. A mismatch in `t` doesn’t satisfy that requirement, so I should only advance `j`. I advance `i` only when I consume a matching character.

🧑‍💼 **Interviewer:**
> What happens when `s` is empty?

🧑‍💻 **Interviewee:**
> Initially `i` is zero and `len(s)` is zero. The loop condition `i < len(s)` is false, and the final check returns `0 == 0`, which is true. That matches the definition: the empty string is a subsequence of every string.

---

## Act 4 — Correctness & Complexity

🧑‍💼 **Interviewer:**
> State your invariant and use it to explain correctness.

🧑‍💻 **Interviewee:**
> After processing the prefix `t[0:j]`, `i` is the number of characters from `s` that I’ve matched in order using that prefix.
>
> For each new character in `t`, if it doesn’t equal the next required character `s[i]`, I leave `i` unchanged and continue. If it matches, I increment `i`, which means the matched prefix of `s` has grown by one while preserving order. The scan stops successfully once `i == len(s)`. If `t` ends first, the final check is false because at least one required character remains unmatched. So the final condition is true exactly when `s` is a subsequence of `t`.

🧑‍💼 **Interviewer:**
> What are the complexity bounds?

🧑‍💻 **Interviewee:**
> `j` moves through `t` at most once, so time is `O(len(t))`. I only keep two integer indexes, so the extra space is `O(1)`. The scan can stop early as soon as all of `s` is matched.

---

## Act 5 — Walkthrough & Edge Cases

🧑‍💼 **Interviewer:**
> Walk through `s = "abc"` and `t = "ahbgdc"`.

🧑‍💻 **Interviewee:**
> I need `a` first. The first character of `t` is `a`, so I advance `i` to the next required character, `b`. I skip `h`, then match `b`. I skip `g`, then match `c`. Now `i == len(s)`, so the result is true.

🧑‍💼 **Interviewer:**
> And what about `s = "axc"` with the same `t`?

🧑‍💻 **Interviewee:**
> I match `a`, then scan the rest of `t` looking for `x`. It never appears, so `i` stays at the position for `x` when the scan ends. Since `i` is less than `len(s)`, I return false.

🧑‍💼 **Interviewer:**
> What boundary cases would you check?

🧑‍💻 **Interviewee:**
> I’d check:
>
> | `s` | `t` | Expected | Why |
> |---|---|---|---|
> | `""` | `"anything"` | `true` | Empty `s` is always a subsequence |
> | `"a"` | `""` | `false` | There is no character to match |
> | `"abc"` | `"abc"` | `true` | Exact match |
> | `"abc"` | `"ahbgdc"` | `true` | Matches appear in order with gaps |
> | `"axc"` | `"ahbgdc"` | `false` | Required `x` never appears |
> | `"aa"` | `"a"` | `false` | One position cannot satisfy two required characters |
>
> The empty `s` and repeated-character cases are especially useful for checking the invariant and confirming that every match consumes a separate position in `t`.

> 💬 **Phrase bank:** *"I’ve matched the prefix…"*, *"The next required character never appears."*, *"Each match consumes a later position in `t`."*

---

## Act 6 — Follow-ups

🧑‍💼 **Interviewer:**
> Suppose we need to check many different strings `s` against the same large `t`. What could you do?

🧑‍💻 **Interviewee:**
> If there are many queries, rescanning all of `t` for each one may be expensive. I could preprocess `t` into a next-occurrence table: for each position and character, record the next index where that character appears. Then each character in a query can jump directly to its next occurrence. The preprocessing takes time and space proportional to `len(t)` times the alphabet size, and each query takes time proportional to its length. For a small fixed alphabet such as lowercase English letters, that can be a useful trade-off.

🧑‍💼 **Interviewer:**
> What changes if the strings can contain arbitrary Unicode characters?

🧑‍💻 **Interviewee:**
> In Go, indexing a string gives a byte, not a Unicode code point. If the specification means matching Unicode code points, I’d iterate over runes or convert the strings to rune slices. Converting to runes uses `O(len(s) + len(t))` extra space; a streaming rune iteration can avoid storing both converted slices, though it needs more careful iteration. I’d first clarify whether matching is defined by bytes, code points, or normalized text.

🧑‍💼 **Interviewer:**
> Is this the same as finding a substring?

🧑‍💻 **Interviewee:**
> No. A substring must be contiguous, while a subsequence may skip characters as long as it preserves order. That’s why this scan keeps moving through `t` after a match instead of requiring the next character in `s` to be adjacent.

---

## Scorecard (Interviewer's internal notes)

| Dimension | Verdict | Comment |
|---|---|---|
| Clarification | 🟢 Strong | Distinguished subsequence from substring and confirmed the input alphabet |
| Pattern recognition | 🟢 Strong | Correctly identified a unidirectional scan |
| Pointer roles | 🟢 Strong | Explained `j` as the scan position and `i` as the matched-prefix length |
| Correctness | 🟢 Strong | Stated an invariant and connected it to the final `i == len(s)` check |
| Edge cases | 🟢 Strong | Covered empty input, missing characters, order, and repeated characters |
| Complexity | 🟢 Strong | Gave `O(len(t))` time and `O(1)` extra space |
| Communication | 🟢 Strong | Explained why mismatches advance only the scan pointer |

---

## Phrase bank for practice

| Situation | English you can reuse |
|---|---|
| Clarifying | "A subsequence preserves order, but the characters don’t have to be adjacent." |
| Defining state | "This index is the next character I need to match." |
| Explaining a mismatch | "This character doesn’t satisfy the current requirement, so I’ll keep scanning." |
| Explaining a match | "I’ve matched one more character from the required sequence." |
| Correctness | "After this prefix of `t`, `i` tells me how much of `s` has matched in order." |
| Edge case | "The empty sequence is matched immediately." |
| Complexity | "I scan each character of `t` at most once." |
