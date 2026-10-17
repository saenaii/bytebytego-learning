# 125. Valid Palindrome

## Quick Review

This problem is best solved with **inward two pointers**.

Start `left` at the beginning of the string and `right` at the end. Each pointer skips non-alphanumeric characters. Then compare the two remaining characters case-insensitively:

- If they differ, the string is not a palindrome.
- If they match, move both pointers toward the center.
- Stop when the pointers meet or cross.

The key invariant is:

> All valid characters outside the current `[left, right]` range have already been checked and form matching pairs.

The accepted approach has:

- **Time:** `O(n)` — each pointer moves across the string at most once.
- **Extra space:** `O(1)` — it does not create a filtered copy.
- **Important detail:** `0` is numeric zero or the NUL byte, while `'0'` is the character `0`. Digit checks must use `b >= '0' && b <= '9'`.
- **Scope:** This implementation handles ASCII letters and digits. It is not a general Unicode case-folding solution.

A concise interview answer is:

> I’ll use two pointers, one at each end of the string. Each pointer skips non-alphanumeric characters, then I compare the remaining characters case-insensitively. If they differ, the string is not a palindrome; otherwise, both pointers move inward. This takes `O(n)` time and `O(1)` extra space.

## Problem

Given a string, determine whether it is a palindrome after converting uppercase letters to lowercase and ignoring all non-alphanumeric characters.

For example:

```text
"A man, a plan, a canal: Panama" -> true
"race a car"                    -> false
" "                             -> true
```

The important detail is that the comparison is case-insensitive and only ASCII letters and digits participate in the comparison.

## Pattern recognition

This is an inward two-pointer problem:

- `left` starts at the beginning of the string.
- `right` starts at the end.
- Non-alphanumeric characters are skipped from either side.
- The remaining characters are compared while the pointers move toward the center.

The palindrome property is symmetric. If the characters at the two ends do not match, the string cannot be a palindrome. If they do match, both ends can be removed from consideration.

## First approach: filter, then compare

The first solution separates the problem into two phases:

1. Build a new string containing only letters and digits.
2. Compare characters from the two ends of the filtered string.

The supplied code follows this shape:

```go
func isPalindrome(s string) bool {
	str := filterString(s)

	length := len(str)
	if length <= 1 {
		return true
	}

	for i, j := 0, length-1; i < j; i, j = i+1, j-1 {
		s1, s2 := str[i], str[j]

		if s1 >= 'A' && s1 <= 'Z' {
			s1 += 32
		}
		if s2 >= 'A' && s2 <= 'Z' {
			s2 += 32
		}

		if s1 != s2 {
			return false
		}
	}
	return true
}

func filterString(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') {
			return r
		}
		return -1
	}, s)
}
```

### Takeaways from the first approach

- **The algorithm is correct because it normalizes the input first.** Once punctuation and spaces are removed, the remaining comparison is a standard inward traversal.
- **Filtering is a useful first decomposition.** It makes the palindrome loop simple and easy to verify.
- **It uses `O(n)` extra space.** `filterString` creates a second string proportional to the input size. The comparison itself is `O(n)` time, and filtering is another `O(n)` pass, so the total remains `O(n)` time.
- **The debug print should not be part of the final solution.** A statement such as `fmt.Println(str)` adds output as a side effect and is unnecessary for the function contract.
- **The case conversion assumes ASCII letters.** Adding `32` converts ASCII uppercase letters to lowercase, but it is not a general Unicode case-folding operation.
- **The code should use `func`, not `unc`.** The sample’s `unc isPalindrome` is a typo in Go.

## Accepted approach: compare in place

The optimized solution keeps the original string and skips irrelevant characters as it scans:

```go
func isPalindrome(s string) bool {
	left, right := 0, len(s)-1

	for left < right {
		for left < right && !isAlphanumeric(s[left]) {
			left++
		}
		for left < right && !isAlphanumeric(s[right]) {
			right--
		}

		if toLowerCase(s[left]) != toLowerCase(s[right]) {
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

func toLowerCase(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
```

### Why this version is better

- **It uses the two-pointer pattern directly.** There is no preprocessing pass and no filtered copy of the string.
- **It uses `O(1)` auxiliary space.** The input string is scanned in place through indexes.
- **Each pointer only moves forward.** `left` moves from the beginning toward the center and `right` moves from the end toward the center. The total work is `O(n)`.
- **The invariant is clear.** Before each comparison, every character outside `[left, right]` has already been validated as a matching pair, and the current pointers are the next characters that might matter.
- **Skipping happens before comparing.** This prevents punctuation or spaces from being treated as meaningful characters.
- **The spelling can be improved.** The sample calls the helper `isAlphanumric`; `isAlphanumeric` is the correctly spelled name. The spelling does not affect the algorithm.

## The `0` versus `'0'` detail

This condition is correct:

```go
b >= '0' && b <= '9'
```

`b` is a `byte`, and `'0'` and `'9'` are character literals. In Go, character literals have numeric values, so the condition checks whether `b` falls inside the ASCII range for the characters `0` through `9`.

Do not replace the character literals with numeric zero:

```go
b >= 0 && b <= 9 // This checks control-byte values, not digit characters.
```

The value `0` represents the NUL byte. It is different from the character `'0'`, whose ASCII value is `48`. Likewise, `'9'` has ASCII value `57`.

| Expression | Meaning |
| --- | --- |
| `0` | Numeric zero / NUL byte value |
| `'0'` | The character zero, ASCII value `48` |
| `b >= '0' && b <= '9'` | `b` is an ASCII digit character |
| `b >= 0 && b <= 9` | `b` has a very small byte value; it is not a digit check |

The same principle applies to letters: compare against `'A'`, `'Z'`, `'a'`, and `'z'`, not against unrelated numeric values.

## Correctness walkthrough

The accepted algorithm maintains this invariant:

> All characters outside the current `[left, right]` interval have already been checked and form matching pairs after ignoring non-alphanumeric characters and normalizing case.

At each iteration:

1. Move `left` rightward until it points to an alphanumeric character or meets `right`.
2. Move `right` leftward until it points to an alphanumeric character or meets `left`.
3. Compare the normalized characters.
4. Return `false` immediately if they differ.
5. Otherwise, move both pointers inward and preserve the invariant.

If the pointers meet or cross, every relevant pair matched, so the string is a palindrome.

## Complexity

| Approach | Time | Extra space | Main trade-off |
| --- | --- | --- | --- |
| Filter then compare | `O(n)` | `O(n)` | Simpler comparison phase, but creates a filtered string |
| Skip while comparing | `O(n)` | `O(1)` | Slightly more pointer logic, but avoids the copy |

Both approaches are linear because each input character is examined only a constant number of times. The second approach is preferred when the goal is constant extra space.

## Edge cases to test

- Empty string: `""` is valid.
- One character: always valid.
- Only punctuation or spaces: valid because no characters remain to compare.
- Mixed case: `"Aa"` is valid.
- Different outer characters: return `false` immediately.
- Digits: `"0P"` is invalid; the digit `0` must be treated as a meaningful alphanumeric character.
- A single digit: `"0"` is valid.
- Punctuation between matching characters: `"a!b!a"` is valid.

## Interview answer template

> I’ll use two pointers, one at each end of the string. Each pointer skips non-alphanumeric characters. I compare the remaining characters case-insensitively; if they differ, the string is not a palindrome. If they match, I move both pointers inward. Each pointer moves across the string at most once, so the time complexity is `O(n)` and the extra space is `O(1)`. In Go, I must compare bytes with character literals such as `'0'` and `'9'`; numeric `0` is not the character `'0'`.
