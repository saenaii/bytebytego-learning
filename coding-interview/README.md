# Coding Interview

This section builds a practical toolkit for algorithm interviews. Each pattern is taught from the idea behind it, the signals that suggest using it, and the reasoning needed to explain and implement it under interview pressure.

## Topic layout

Each pattern has its own folder:

```text
coding-interview/
└── <topic>/
    ├── README.md       # The pattern, recognition signals, and solving process
    └── <problem>.md    # One problem and its interview-oriented walkthrough
```

The topic README comes first. Problem notes can then apply the pattern to a concrete question without repeating the complete introduction.

## Topics

| Topic | What it covers |
| --- | --- |
| [Two Pointers](./two-pointers/) | Using two positions in a linear data structure to compare, scan, or transform values while exploiting predictable movement. |

## Recommended study loop

1. Read the topic README until you can explain the pattern without looking at the notes.
2. Practice identifying the pattern from the input shape and the required output.
3. Solve each problem by stating the invariant before writing code.
4. Explain pointer movement, correctness, complexity, and edge cases aloud.
