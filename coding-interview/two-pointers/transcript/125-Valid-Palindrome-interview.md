# 125. Valid Palindrome — Mock Interview Transcript

> English-speaking practice transcript for [125. Valid Palindrome](../125-Valid-Palindrome.md).
> Both roles are written out so the phrasing can be reused in a real interview.

**Format:** Full transcript (Interviewer 🧑‍💼 / Interviewee 🧑‍💻)
**Focus:** Natural interview English — clarification, approach, code, complexity, edge cases, follow-ups

---

## Act 1 — Clarification

🧑‍💼 **Interviewer:**
> Let's start simple. Given a string `s`, return `true` if it is a palindrome, considering only alphanumeric characters and ignoring case. Before you write anything — what questions do you have for me?

🧑‍💻 **Interviewee:**
> Sure, a few things I want to confirm.
>
> First, by "alphanumeric" — do you mean ASCII letters and digits only, or should I also handle Unicode letters, like accented characters?
>
> Second, on case: can I assume a simple ASCII lowercase conversion is enough, or do you expect full Unicode case folding?
>
> Third, is the input guaranteed to be a valid string, or should I handle `null`?
>
> And last — are we optimizing for extra space, or is an `O(n)` filtered copy acceptable?

🧑‍💼 **Interviewer:**
> Good questions. Assume ASCII letters and digits only, simple case-insensitive comparison. The input is a valid string — no `null`. And I do care about space, so try to avoid building a second string.

🧑‍💻 **Interviewee:**
> Got it. So the contract is: ignore everything that isn't `a–z`, `A–Z`, or `0–9`, compare case-insensitively, and do it without extra memory.

> 💬 **Why this act matters in English:** the highest-value opening phrases are *"I want to confirm…"*, *"Do you mean X, or Y?"*, *"Can I assume…?"*. Interviewers read this as seniority, not hesitation.

---

## Act 2 — Approach

🧑‍💼 **Interviewer:**
> Alright. Talk me through your approach before coding.

🧑‍💻 **Interviewee:**
> I'd use **inward two pointers**.
>
> The key observation is that a palindrome is **symmetric**: the first meaningful character must match the last one, the second must match the second-to-last, and so on. So I don't need to compare every pair — I only need to compare from the outside in.
>
> Concretely:
> - `left` starts at index `0`, `right` starts at `len(s) - 1`.
> - Each pointer **skips** characters that aren't alphanumeric.
> - Then I compare the two characters, lowercased.
> - If they differ, I return `false` immediately.
> - If they match, I move both pointers inward.
> - When the pointers meet or cross, everything matched, so I return `true`.
>
> The invariant I'm maintaining is: **every meaningful character outside the current `[left, right]` window has already been checked and forms a matching pair.**

🧑‍💼 **Interviewer:**
> You called it "inward two pointers." Why not a sliding window? Both use a left and right boundary.

🧑‍💻 **Interviewee:**
> Good distinction. A **sliding window** maintains a *contiguous* range whose validity can change as it expands or shrinks — the classic shape is "find the longest substring satisfying some condition," and the window usually grows on the right and shrinks on the left.
>
> Here there's no validity condition on a range, and the two pointers never move in the same direction. They both move **toward the center**, and each move permanently discards the characters behind it. That's inward traversal, not a window.
>
> I'd also distinguish it from **unidirectional two pointers**, like the "move zeroes" problem, where a read pointer and a write pointer both move forward. And from **fast/slow pointers**, where the two pointers move at different speeds to detect a cycle.

🧑‍💼 **Interviewer:**
> And what about the brute-force version?

🧑‍💻 **Interviewee:**
> The most straightforward version is **filter, then compare**: build a new string with only the alphanumeric characters, then check whether that string reads the same forwards and backwards.
>
> It's correct and easy to reason about, but it costs `O(n)` extra space for the copy. Since you said you care about space, I'd rather skip in place and keep it `O(1)`.

> 💬 **Phrase bank:** *"The key observation is…"*, *"Concretely…"*, *"The invariant I'm maintaining is…"*, *"I'd distinguish it from…"*

---

## Act 3 — Coding

🧑‍💼 **Interviewer:**
> Let's see it in Go.

🧑‍💻 **Interviewee:**
> I'll write two small helpers so the main loop stays readable.

```go
func isPalindrome(s string) bool {
	left, right := 0, len(s)-1

	for left < right {
		// skip non-alphanumeric characters from the left
		for left < right && !isAlphanumeric(s[left]) {
			left++
		}
		// skip non-alphanumeric characters from the right
		for left < right && !isAlphanumeric(s[right]) {
			right--
		}

		if toLower(s[left]) != toLower(s[right]) {
			return false
		}

		left++
		right--
	}
	return true
}

func isAlphanumeric(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z')
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
```

> One detail worth calling out: I'm indexing `s[left]` as a `byte`, which is fine because we agreed on ASCII-only input. If we needed Unicode, I'd have to iterate over `rune`s instead.

🧑‍💼 **Interviewer:**
> Look at `b >= '0' && b <= '9'`. What happens if you write `b >= 0 && b <= 9` instead?

🧑‍💻 **Interviewee:**
> That would be a bug. `0` is the numeric value zero — the NUL byte. `'0'` is the *character* zero, which has ASCII value `48`, and `'9'` is `57`.
>
> So `b >= 0 && b <= 9` only matches control bytes like tab and newline, and it would reject every actual digit. The character-literal version is the correct digit check.

---

## Act 4 — Correctness & Complexity

🧑‍💼 **Interviewer:**
> Convince me this never skips a valid answer.

🧑‍💻 **Interviewee:**
> I'll argue by the invariant. At the top of every iteration, all meaningful characters strictly outside `[left, right]` have already been compared and matched. That's true initially because the range is empty outside.
>
> Inside the loop, the skip steps only move pointers **past characters that can't participate** in the comparison, so they can't skip a meaningful character. Then I compare the next two meaningful characters. If they differ, the string is definitively not a palindrome, because those two characters must be a mirror pair. If they match, I move both pointers inward, and the invariant is restored for a strictly smaller range.
>
> When the pointers meet or cross, no meaningful pair is left uncompared, so returning `true` is correct.

🧑‍💼 **Interviewer:**
> Complexity?

🧑‍💻 **Interviewee:**
> `O(n)` time and `O(1)` extra space.
>
> The time bound isn't `O(n²)` despite the nested loops: `left` only ever increases and `right` only ever decreases, so **each index is visited at most once by each pointer**. The total number of pointer moves is bounded by `n`. Space is constant because I only keep two integer indices — no filtered copy.

---

## Act 5 — Edge Cases

🧑‍💼 **Interviewer:**
> What would you test?

🧑‍💻 **Interviewee:**
> Let me walk through the boundary cases:
>
> | Input | Expected | Why |
> |---|---|---|
> | `""` | `true` | Empty string is trivially a palindrome |
> | `"a"` | `true` | Single character, pointers meet immediately |
> | `" "` | `true` | No meaningful characters remain |
> | `".,!?"` | `true` | Same — everything is skipped |
> | `"Aa"` | `true` | Case-insensitive comparison |
> | `"0P"` | `false` | The digit `0` is meaningful; don't skip it |
> | `"0"` | `true` | A lone digit is valid |
> | `"a!b!a"` | `true` | Punctuation between matching characters |
> | `"ab"` | `false` | Fails on the first comparison |
>
> The two cases I'd be most careful about are `"0P"` — where a buggy digit check would silently skip the `0` — and the all-punctuation case, where both pointers run past each other.

> 💬 **Phrase bank:** *"Let me walk through the boundary cases."*, *"The cases I'd be most careful about are…"*

---

## Act 6 — Follow-ups

🧑‍💼 **Interviewer:**
> Nice. Now suppose the input can contain Unicode letters — `"été"` should be valid. What breaks?

🧑‍💻 **Interviewee:**
> Two things break.
>
> First, indexing `s[i]` gives me a `byte`, not a character. In Go, `é` is two bytes in UTF-8, so I'd slice a multi-byte character in half. I need to iterate over `rune`s — either by converting to `[]rune` once, or by decoding with `utf8.DecodeRuneInString` from both ends.
>
> Second, my hand-rolled ASCII checks are wrong. I'd use `unicode.IsLetter`, `unicode.IsDigit`, and `unicode.ToLower`.
>
> Note the trade-off: converting to `[]rune` costs `O(n)` space, so I'd lose the constant-space property. If I decode runes in place from both ends, I can keep it `O(1)` extra space, at the cost of more fiddly pointer arithmetic.

🧑‍💼 **Interviewer:**
> Last one. Same problem, but the input arrives as a stream and you can't store it. What changes?

🧑‍💻 **Interviewee:**
> Then a single forward pass can't decide, because the decision depends on characters at the far end.
>
> The practical options are:
> - **Buffer everything**, then run the two-pointer scan — back to `O(n)` space.
> - **Keep a stack** of the first half and pop it against the second half — also `O(n)` space, but it works incrementally and gives early failure as soon as a mismatch appears.
> - If the stream is a **singly linked list**, I can't walk backwards, so I'd **reverse the second half** in place, compare, then restore it — `O(1)` space and `O(n)` time.
>
> So the algorithm doesn't change conceptually; the data structure determines how I get the "right" pointer.

🧑‍💼 **Interviewer:**
> And a variation: allow **at most one character deletion**. How would you adapt it?

🧑‍💻 **Interviewee:**
> That's "Valid Palindrome II". The two-pointer scan stays, but when I find a mismatch, I get a choice: delete the left character or delete the right character. So I try both — check whether `s[left+1..right]` or `s[left..right-1]` is a palindrome — and return `true` if either works.
>
> It's still `O(n)` time, because each of those sub-checks is a linear scan and I do at most two of them. Space stays `O(1)`.

---

## Scorecard (Interviewer's internal notes)

| Dimension | Verdict | Comment |
|---|---|---|
| Clarification | 🟢 Strong | Asked about ASCII vs Unicode, case folding, null, and space budget up front |
| Pattern recognition | 🟢 Strong | Named inward two pointers and contrasted it with window / unidirectional / fast-slow |
| Invariant statement | 🟢 Strong | Stated the outside-inside invariant before writing code |
| Code quality | 🟢 Strong | Helpers extracted, no debug prints, no filtered copy |
| Complexity analysis | 🟢 Strong | Explained why nested loops are still linear |
| Edge cases | 🟢 Strong | Covered `"0P"` and all-punctuation cases |
| Communication | 🟡 Acceptable | Slightly verbose in Act 6 — could lead with the answer, then justify |

---

## Phrase bank for practice

| Situation | English you can reuse |
|---|---|
| Clarifying | "I want to confirm one thing before I start." / "Do you mean X, or Y?" / "Can I assume…?" |
| Proposing | "The key observation is…" / "I'd approach this with…" / "Concretely, here's the plan." |
| Justifying | "The invariant I'm maintaining is…" / "This is safe because…" |
| Distinguishing | "I'd distinguish this from…" / "This is similar to X, but the difference is…" |
| Complexity | "The total work is bounded by n because…" / "Space is constant since I only keep two indices." |
| Hedging gracefully | "I'd start with the simple version, then optimize if needed." / "That's a trade-off I'd be happy to discuss." |
| Buying time | "Let me think about that for a second." / "Let me walk through a concrete example." |
