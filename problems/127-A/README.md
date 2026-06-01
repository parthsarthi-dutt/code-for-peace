## Problem
               
You have $n$ beads, and each bead has a specific color represented by an integer $a_i$. 

You want to string all these beads onto necklaces. However, there is a strict rule: **no single necklace can contain two beads of the same color**. 

### Task
**Calculate the minimum number of necklaces required to use up all $n$ beads.**

### Input
The first line contains an integer $t$ ($1 \le t \le 1000$) — the number of test cases.
The first line of each test case contains an integer $n$ ($1 \le n \le 10^5$) — the number of beads.
The second line of each test case contains $n$ integers $a_1, a_2, \ldots, a_n$ ($1 \le a_i \le 10^9$) — the colors of the beads.

### Output
For each test case, print a single integer representing the minimum number of necklaces needed.

### Constraints
The sum of $n$ over all test cases does not exceed $2 \cdot 10^5$.