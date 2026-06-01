# Editorial: 123-A

## Approach

This problem is a variant of the **0/1 Knapsack** problem combined with a number-theory-based preprocessing step.

### Key Observations

1. **Preprocessing**: For each element `v1[i]`, we need to compute the minimum number of operations to reduce it to 1. The allowed operation is: pick a divisor `d` of the current number, and reduce the number by `current / d`. This is computed using BFS/DP over all numbers up to 1000.

2. **DP Formulation**: After preprocessing, we have:
   - `v2[i]` = value of the i-th item
   - `v3[i]` = cost (number of operations) for the i-th item
   - `k` = total budget

   This is a classic 0/1 knapsack: maximize the total value while keeping total cost ≤ k.

### Time Complexity

- **Preprocessing**: `O(N * sqrt(N))` for all divisors
- **DP**: `O(n * k)` per test case

### Space Complexity

- `O(n * k)` for the DP table

## Solution (C++)

```cpp
#include <bits/stdc++.h>
using namespace std;
typedef long long ll;

vector<ll> dp2(1001, 1e18);

void preprocess() {
    dp2[1] = 0;
    for (ll i = 1; i <= 1000; i++) {
        dp2[i] = min(dp2[i], i);
        for (ll j = 1; j <= i; j++) {
            if (i + (i / j) > 1000) continue;
            dp2[i + (i / j)] = min(dp2[i + (i / j)], dp2[i] + 1);
        }
    }
}

ll dprec(ll i, ll sum, vector<vector<ll>>& dp, ll k, vector<ll>& v2, vector<ll>& v3) {
    if (i == v3.size()) return 0;
    if (dp[i][sum] != -1) return dp[i][sum];
    ll res = dprec(i + 1, sum, dp, k, v2, v3);  // skip
    if (sum + v3[i] <= k)
        res = max(res, dprec(i + 1, sum + v3[i], dp, k, v2, v3) + v2[i]);  // take
    return dp[i][sum] = res;
}

void solve(ll n) {
    ll k;
    cin >> k;
    vector<ll> v1(n), v2(n), v3(n);
    for (auto& x : v1) cin >> x;
    for (auto& x : v2) cin >> x;
    for (ll i = 0; i < n; i++) v3[i] = dp2[v1[i]];
    vector<vector<ll>> dp(n + 1, vector<ll>(min(k + 1, 12 * n + 1) + 2, -1));
    cout << dprec(0, 0, dp, k, v2, v3) << "\n";
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    preprocess();
    ll t;
    cin >> t;
    while (t--) {
        ll n;
        cin >> n;
        solve(n);
    }
}
```
