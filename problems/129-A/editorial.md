# Editorial: 128-C (Absolute Alignment)

## Approach

This problem requires understanding the properties of modular arithmetic and the Greatest Common Divisor (GCD).

### Key Observations
1. If $a \pmod M = b \pmod M$, then their difference must be a multiple of $M$. Therefore, $M$ must divide $|a - b|$.
2. This logic extends to the entire array: $M$ must divide the absolute difference between every pair of elements.
3. To maximize $M$, $M$ should be the Greatest Common Divisor (GCD) of all adjacent differences in the array: $M = \gcd(|a_2 - a_1|, |a_3 - a_2|, \ldots, |a_n - a_{n-1}|)$.
4. **Edge Case**: If all elements in the array are exactly the same, their differences are all `0`. Since any integer divides `0`, $M$ can be infinitely large. In this case, we should output `0`.

### Time Complexity
- $O(N \log(\min(A)))$ per test case due to the Euclidean algorithm for GCD. Overall $O(\sum N \log(\min(A)))$.

### Space Complexity
- $O(1)$ auxiliary space.

## Solution (C++)

```cpp
#include <bits/stdc++.h>
using namespace std;
typedef long long ll;

ll gcd(ll a, ll b) {
    if (b == 0) return a;
    return gcd(b, a % b);
}

void solve() {
    int n;
    cin >> n;
    vector<ll> a(n);
    for (int i = 0; i < n; i++) cin >> a[i];

    ll ans = 0;
    for (int i = 1; i < n; i++) {
        ans = gcd(ans, abs(a[i] - a[i - 1]));
    }
    
    cout << ans << "\n";
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t; cin >> t;
    while(t--) solve();
}