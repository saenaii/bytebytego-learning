# Two Pointers

## The takeaway

The two-pointer pattern uses two variables to represent positions in a linear data structure, such as an array, string, or linked list. The pointers let us compare or coordinate two positions while using a predictable property of the input to decide how they move.

The main interview insight is this: if the input gives us enough structure to rule out many candidates at once, we can often replace a nested pairwise scan with a single coordinated pass. That commonly reduces time complexity from `O(n²)` to `O(n)`, although the improvement depends on the input and the pointer invariant.

Two pointers are not a single algorithm. They are a way to express a search or transformation in which each pointer has a clear responsibility and every movement can be justified.

## What is a pointer?

A pointer here is simply a variable holding an index or position. It does not require a language-level pointer type.

For an array, pointers might be integer indexes:

```text
nums = [1, 2, 4, 7, 11]
left = 0
right = 4
```

`left` refers to `nums[0]`, while `right` refers to `nums[4]`. The algorithm compares those positions and moves one or both pointers according to a rule.

## Why two pointers help

A straightforward pair-search often compares every pair:

```text
for i from 0 to n - 1:
    for j from i + 1 to n - 1:
        compare(nums[i], nums[j])
```

This examines `O(n²)` pairs. If the data is sorted, symmetric, or otherwise predictable, one comparison can tell us that an entire group of pairs cannot work. We can then move a pointer past that group instead of checking every pair individually.

The optimization is valid only when the pointer movement follows an invariant. For example, in an ascending array, moving a pointer to the right never decreases the value it references. That predictable change is what lets the algorithm discard candidates safely.

## Three common strategies

| Strategy | Initial positions | Pointer responsibilities | Good fit |
| --- | --- | --- | --- |
| Inward traversal | Opposite ends | Compare both ends and move one or both toward the center | Pair sums in sorted arrays, palindrome checks, maximum container problems |
| Unidirectional traversal | Usually the same end | One pointer explores; the other tracks a boundary or destination | Moving zeroes, filtering values in place, merging or compacting data |
| Staged traversal | One pointer scans first | The second pointer starts work after the first finds a qualifying position | Problems with a search phase followed by a follow-up scan |

### Inward traversal

Start `left` at the beginning and `right` at the end. Compare the two values, then move the pointer that can improve the result or eliminate impossible candidates.

```text
left = 0
right = n - 1

while left < right:
    current = combine(nums[left], nums[right])

    if current satisfies the goal:
        record or return the answer
    else if current is too small:
        left += 1
    else:
        right -= 1
```

The important part is not the variable names. It is the reason a move is safe. In a sorted pair-sum problem, if the sum is too small, moving `right` left would make the sum even smaller, so `left` must move right.

### Unidirectional traversal

Both pointers move in the same direction, but they do different jobs. A fast or read pointer explores the input; a slow or write pointer records the next position for an accepted value.

```text
write = 0

for read from 0 to n - 1:
    if nums[read] should be kept:
        nums[write] = nums[read]
        write += 1
```

This is useful when the result can be built in place. The region before `write` has a maintained meaning, such as “all values kept so far.”

### Staged traversal

One pointer first finds a position or condition of interest. A second pointer then performs additional work relative to that position.

```text
first = find_starting_position()

if first exists:
    second = first
    while second satisfies the follow_up_condition:
        process(second)
        second += 1
```

The two pointers may not move at the same time. The key is that each stage has a distinct purpose and the second traversal does not redo work that the first stage already ruled out.

## When to look for two pointers

Two pointers are especially promising when:

- The input is a linear structure: an array, string, or linked list.
- The input is sorted, or moving in one direction gives a predictable value change.
- The problem asks about a pair, two ends, or a result formed from two values.
- The input has symmetry, such as a palindrome.
- The output can be built by keeping a boundary between processed and unprocessed data.
- The problem asks for an in-place transformation and extra memory should be avoided.

These are signals, not proof. Always establish an invariant before choosing the pattern.

## Interview-solving process

Use this sequence when explaining a two-pointer solution:

1. **Identify the invariant or predictable property.** Explain what the input structure tells us about a pointer move.
2. **Define each pointer.** State exactly what `left`, `right`, `read`, `write`, or another pointer represents.
3. **Explain movement.** For every branch, say why moving that pointer is safe and what candidates are eliminated.
4. **State the stopping condition.** Usually the pointers meet, cross, or the scan reaches the end.
5. **Prove that no valid answer is skipped.** Connect each discarded region to the invariant.
6. **Analyze complexity.** Count total pointer movements rather than multiplying nested-looking sections automatically. If each pointer moves at most `n` times, the total is often `O(n)` time; state any sorting cost separately.
7. **Test edge cases.** Consider empty input, one element, duplicate values, already-valid input, no solution, and pointer meeting or crossing.

## A quick pattern-selection guide

| If the problem looks like... | Start by asking... |
| --- | --- |
| Find two values in a sorted array | Which pointer movement makes the target larger or smaller? |
| Compare both ends of a sequence | Can the outer pair be accepted or rejected before moving inward? |
| Remove or keep values in place | What does the region before the write pointer guarantee? |
| Find a pair of boundaries or a range | Does the right pointer expand the candidate and the left pointer shrink it? If so, this may be sliding window instead. |
| Detect a cycle or midpoint in a linked list | Do pointer speeds differ? If so, this may be fast/slow pointers instead. |

## Related patterns

- **Fast and slow pointers:** A specialized two-pointer technique in which pointers move at different speeds, often to detect cycles or find a linked-list midpoint.
- **Sliding window:** Maintains a contiguous interval, usually with a left boundary and a right boundary. It is related to unidirectional two pointers, but the window’s validity condition and resizing behavior are the central idea.
- **Binary search:** Uses a shrinking search interval and often has two boundaries, but it chooses a midpoint and discards half the search space rather than coordinating two data positions directly.

## Real-world intuition: memory compaction

As an analogy, a garbage collector can compact memory by using one pointer to scan for live objects and another pointer to track the next free destination. The scan pointer moves through the heap; the destination pointer advances only when a live object needs to be kept. This is the same “explore with one position, maintain a boundary with another” idea as an in-place array transformation.

The analogy is limited: production garbage collectors have additional concerns such as object references, allocation policy, pauses, and concurrency. It illustrates the pointer roles, not a complete garbage-collection design.

## Related problems

- [125. Valid Palindrome](./125-Valid-Palindrome.md) — skip non-alphanumeric characters while comparing matching positions from both ends.
- [11. Container With Most Water](./11-container-with-most-water.md) — maximize area by moving the shorter boundary inward.
- [167. Two Sum II — Input Array Is Sorted](./167-two-sum2-input-array-is-sorted.md) — use sorted order to move two pointers toward a target pair.

Future problem walkthroughs should explain the recognition signal, invariant, pointer movement, proof of correctness, complexity, edge cases, and implementation approach.

## Source and further reading

This note is an original study guide informed by [ByteByteGo’s Introduction to Two Pointers](https://bytebytego.com/courses/coding-patterns/two-pointers/introduction-to-two-pointers). The source lesson organizes the pattern around inward, unidirectional, and staged traversal and connects the pattern to problems such as pair sum, triplet sum, palindrome validation, moving zeroes, and next lexicographical sequence.
