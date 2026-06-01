## Problem
               
You are given a binary string $s$ of length $n$. 

### Operation
In one operation, you can choose an index $i$ ($1 \le i < |s|$) such that $s_i \neq s_{i+1}$. You then remove both $s_i$ and $s_{i+1}$ from the string. The remaining parts of the string are concatenated. 

### Given
You apply operations until it is impossible to apply any more. A sequence of chosen indices $t_1, t_2, \ldots, t_k$ represents an elimination process. Two elimination processes are considered different if the sequences of chosen indices are different. 

### Task
**Calculate the number of different valid elimination processes that result in the shortest possible final string modulo 998244353.**

### Input
The first line contains an integer $t$ ($1 \le t \le 1000$) — the number of test cases.
The first line of each test case contains an integer $n$ ($1 \le n \le 5000$) — the length of the binary string.
The second line of each test case contains the binary string $s$ of length $n$.

### Output
For each test case print one integer: the number of valid sequences of operations modulo 998244353.

### Constraints
The sum of $n^2$ among all test cases is guaranteed to not exceed $2.5 \cdot 10^7$.