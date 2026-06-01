# Editorial: 128-A (Optimal Pairing)

## Approach

This problem relies on a greedy algorithm combined with sorting.

### Key Observations
1. To minimize the maximum sum of any pair, we want to balance the pairs as much as possible.
2. If we pair the largest element with anything other than the smallest element, we risk creating a needlessly large sum.
3. The optimal greedy strategy is to sort the array and pair the $i$-th smallest element with the $i$-th largest element.
4. Specifically, pair $a[i]$ with $a[2n - 1 - i]$ for all $0 \le i < n$. The answer is the maximum sum among these $n$ pairs.

### Time Complexity
- $O(N \log N)$ per test case due to sorting. Overall $O(\sum N \log N)$.

### Space Complexity
- $O(1)$ auxiliary space if sorting in place.

## Solution (C++)

```cpp
#include <bits/stdc++.h>
using namespace std;
typedef long long ll;

void solve() {
    int n;
    cin >> n;
    vector<ll> a(2 * n);
    for (int i = 0; i < 2 * n; i++) cin >> a[i];
    
    sort(a.begin(), a.end());
    
    ll max_sum = 0;
    for (int i = 0; i < n; i++) {
        max_sum = max(max_sum, a[i] + a[2 * n - 1 - i]);
    }
    cout << max_sum << "\n";
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t; cin >> t;
    while(t--) solve();
}