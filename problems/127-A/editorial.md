# Editorial: 127-A (Minimum Necklaces)

## Approach

This is a fundamental frequency counting (hashing) problem. 

### Key Observations
1. If a color appears $k$ times in the array, you absolutely must have at least $k$ separate necklaces, because no two beads of the same color can share a necklace.
2. Conversely, if you have $k$ necklaces (where $k$ is the maximum frequency of any color), you can easily distribute the remaining beads across these $k$ necklaces without violating the rule.
3. Therefore, the answer is simply the maximum frequency of any single element in the array.

### Time Complexity
- $O(N)$ per test case using a hash map (or $O(N \log N)$ if sorting/using a tree map). Overall $O(\sum N)$.

### Space Complexity
- $O(N)$ for the hash map storing the frequencies.

## Solution (C++)

```cpp
#include <bits/stdc++.h>
using namespace std;

void solve() {
    int n;
    cin >> n;
    map<int, int> freq;
    int max_freq = 0;
    for (int i = 0; i < n; i++) {
        int color;
        cin >> color;
        freq[color]++;
        max_freq = max(max_freq, freq[color]);
    }
    cout << max_freq << "\n";
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t; cin >> t;
    while(t--) solve();
}