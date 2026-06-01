## Problem
               
You are given an array $a$ of $n$ positive integers. You can perform the following operation any number of times (possibly zero):

### Operation
Choose an index $i$ ($1 \le i \le n$) and multiply $a_i$ by 2. 

### Given
You need to make the array satisfy the condition: $a_i \ge a_{i-1}$ for all $2 \le i \le n$. 
Furthermore, you want to minimize the total number of operations performed.

### Task
**Calculate the minimum number of operations required to make the array non-decreasing.**

### Input
The first line contains an integer $t$ ($1 \le t \le 10^4$) — the number of test cases.
The first line of each test case contains a single integer $n$ ($1 \le n \le 10^5$) — the length of the array $a$.
The second line of each test case contains $n$ integers $a_1, a_2, \ldots, a_n$ ($1 \le a_i \le 10^9$) — the elements of the array.

### Output
For each test case, output a single integer — the minimum number of operations required.

### Constraints
The sum of $n$ over all test cases does not exceed $2 \cdot 10^5$.