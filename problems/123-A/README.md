## Problem
               
We start with a permutation `a1, a2, ..., an` and with an empty array `b`. We apply the following operation `k` times.

### Operation

On the `i`-th iteration, we select an index `ti` (1 ≤ ti ≤ n - i + 1), remove `ati` from the array, and append one of the numbers `ati-1` or `ati+1` (if they are within the array bounds) to the right end of the array `b`. Then we move elements `ati+1, ..., an` to the left in order to fill in the empty space.

### Given

You are given the initial permutation `a1, a2, ..., an` and the resulting array `b1, b2, ..., bk`. All elements of an array `b` are distinct.

### Task

**Calculate the number of possible sequences of indices `t1, t2, ..., tk` modulo 998244353.**

### Input

Each test contains multiple test cases. The first line contains an integer `t` (1 ≤ t ≤ 100000), denoting the number of test cases, followed by a description of the test cases.

The first line of each test case contains two integers `n`, `k` (1 ≤ k < n ≤ 200000): sizes of arrays `a` and `b`.

The second line of each test case contains `n` integers `a1, a2, ..., an` (1 ≤ ai ≤ n): elements of `a`. All elements of `a` are distinct.

The third line of each test case contains `k` integers `b1, b2, ..., bk` (1 ≤ bi ≤ n): elements of `b`. All elements of `b` are distinct.

### Output

For each test case print one integer: the number of possible sequences modulo 998244353.

### Constraints

The sum of all `n` among all test cases is guaranteed to not exceed **200000**.