## Problem
               
You are given an array $a$ of $n$ positive integers. 

You want to find the largest positive integer $M$ such that all elements in the array leave the exact same remainder when divided by $M$. In other words, $a_1 \pmod M = a_2 \pmod M = \ldots = a_n \pmod M$.

If $M$ can be infinitely large, output `0`.

### Task
**Calculate the maximum possible value for $M$. If it can be arbitrarily large, output 0.**

### Input
The first line contains an integer $t$ ($1 \le t \le 10^4$) — the number of test cases.
The first line of each test case contains a single integer $n$ ($2 \le n \le 10^5$).
The second line contains $n$ integers $a_1, a_2, \ldots, a_n$ ($1 \le a_i \le 10^{18}$). Note that elements can be very large.

### Output
For each test case, output a single integer representing the maximum $M$, or `0` if $M$ is unbounded.

### Constraints
The sum of $n$ over all test cases does not exceed $2 \cdot 10^5$.