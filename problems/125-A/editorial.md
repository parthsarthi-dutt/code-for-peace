# Editorial: 125-A (Prefix Equality)

## Approach

This problem requires a greedy approach. To minimize operations, we should only multiply an element by 2 when it is strictly necessary to make it at least as large as the previous element. 

### Key Observations

1. **Greedy Choice**: If $a_i < a_{i-1}$, we must multiply $a_i$ by 2 until $a_i \ge a_{i-1}$. 
2. **Avoiding Overflow**: If we actually multiply the numbers, they can grow up to $10^9 \cdot 2^{10^5}$, which will easily overflow any standard data type. 
3. **Logarithmic State**: Instead of maintaining the actual values, we can maintain the *number of operations* applied to each element. Let $cnt_i$ be the number of times $a_i$ is multiplied by 2. 
We need $a_i \cdot 2^{cnt_i} \ge a_{i-1} \cdot 2^{cnt_{i-1}}$.
Taking base-2 logarithms is an option, but floating-point precision issues will fail edge cases. Instead, we can compare the base numbers directly by shifting.

We can compute the required shifts:
If $a_i < a_{i-1}$, we find the minimum $k$ such that $a_i \cdot 2^k \ge a_{i-1}$. Then $cnt_i = cnt_{i-1} + k$.
If $a_i \ge a_{i-1}$, we might be able to *reduce* the multiplier. We find the maximum $k$ such that $a_i \ge a_{i-1} \cdot 2^k$. Then $cnt_i = \max(0, cnt_{i-1} - k)$.

### Time Complexity
- $O(N)$ per test case, as computing the required shifts takes $O(1)$ amortized time or using bitwise operations/`__builtin_clz`. Overall $O(\sum N)$.

### Space Complexity
- $O(1)$ auxiliary space, or $O(N)$ if storing the operations array.

## Solution (C++)

```cpp
#include <bits/stdc++.h>
using namespace std;
typedef long long ll;

void solve() {
    int n;
    cin >> n;
    vector<ll> a(n);
    for (int i = 0; i < n; i++) cin >> a[i];

    ll ans = 0;
    ll prev_ops = 0;

    for (int i = 1; i < n; i++) {
        ll curr_ops = 0;
        if (a[i] < a[i-1]) {
            ll temp = a[i];
            ll req = 0;
            while (temp < a[i-1]) {
                temp *= 2;
                req++;
            }
            curr_ops = prev_ops + req;
        } else {
            ll temp = a[i];
            ll reduce = 0;
            while (temp / 2 >= a[i-1]) {
                temp /= 2;
                reduce++;
            }
            curr_ops = max(0LL, prev_ops - reduce);
        }
        ans += curr_ops;
        prev_ops = curr_ops;
    }
    cout << ans << "\n";
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t; cin >> t;
    while(t--) solve();
}