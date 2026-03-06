// Author - Parthsarthi Dutt

#pragma GCC optimize("O3,unroll-loops")

#include <bits/stdc++.h>
#include <ext/pb_ds/assoc_container.hpp>
#include <ext/pb_ds/tree_policy.hpp>

using namespace std;
using namespace chrono;
using namespace __gnu_pbds;

typedef long long ll;
//       _ _
//      (•_•)
//     (   (>   Idk how this works
//      /  |   but it does
//

#ifndef ONLINE_JUDGE
#define debarr(a, n)     \
    cerr << #a << " : "; \
    F(i, n)              \
    cerr << a[i] << " "; \
    cerr << endl;
#define debmat(mat, r, c)         \
    cerr << #mat << " :\n";       \
    F(i, r)                       \
    {                             \
        F(j, c)                   \
        cerr << mat[i][j] << " "; \
        cerr << endl;             \
    }
#define pr(...) dbs(#__VA_ARGS__, __VA_ARGS__)
template <class S, class T>
ostream &operator<<(ostream &os, const pair<S, T> &p) { return os << "(" << p.first << "," << p.second << ")"; }
template <class T>
ostream &operator<<(ostream &os, const vector<T> &v)
{
    os << "[ ";
    for (auto &i : v)
        os << i << " ";
    return os << "]";
}
template <class T>
ostream &operator<<(ostream &os, const set<T> &s)
{
    os << "[ ";
    for (auto &i : s)
        os << i << " ";
    return os << "]";
}
template <class T>
ostream &operator<<(ostream &os, const multiset<T> &s)
{
    os << "[ ";
    for (auto &i : s)
        os << i << " ";
    return os << "]";
}
template <class S, class T>
ostream &operator<<(ostream &os, const map<S, T> &m)
{
    os << "[ ";
    for (auto &i : m)
        os << i << " ";
    return os << "]";
}
template <class T>
ostream &operator<<(ostream &os, const unordered_set<T> &s)
{
    os << "[ ";
    for (auto &i : s)
        os << i << " ";
    return os << "]";
}
template <class S, class T>
ostream &operator<<(ostream &os, const unordered_map<S, T> &m)
{
    os << "[ ";
    for (auto &i : m)
        os << i << " ";
    return os << "]";
}
template <class T>
void dbs(string str, T t) { cerr << str << " : " << t << "\n"; }
template <class T, class... S>
void dbs(string str, T t, S... s)
{
    int idx = str.find(',');
    cerr << str.substr(0, idx) << " : " << t << ",";
    dbs(str.substr(idx + 1), s...);
}
template <class T>
void prc(T a, T b)
{
    cerr << "[";
    for (T i = a; i != b; ++i)
    {
        if (i != a)
            cerr << ", ";
        cerr << *i;
    }
    cerr << "]\n";
}
#else
#define pr(...) \
    {           \
    }
#define debarr(a, n) \
    {                \
    }
#define debmat(mat, r, c) \
    {                     \
    }
#endif

#define fastio()                      \
    ios_base::sync_with_stdio(false); \
    cin.tie(NULL);                    \
    cout.tie(NULL)
#define MOD 1000000007
#define MOD1 998244353
#define INF 1e18
#define nline "\n"
#define pb push_back
#define ppb pop_back
#define mp make_pair
#define ff first
#define ss second
#define PI 3.141592653589793238462
#define set_bits __builtin_popcountll
#define sz(x) ((int)(x).size())
#define all(x) (x).begin(), (x).end()
#define input(a)      \
    for (auto &i : a) \
        cin >> i;
#define srt(v) sort(v.begin(), v.end())

typedef unsigned long long ull;
typedef long double lld;
typedef __int128 ell;
typedef tree<pair<ll, ll>, null_type, less<pair<ll, ll>>, rb_tree_tag, tree_order_statistics_node_update> pbds;
typedef pair<ll, ll> pll;
ll gcd(ll a, ll b)
{
    if (b > a)
    {
        return gcd(b, a);
    }
    if (b == 0)
    {
        return a;
    }
    return gcd(b, a % b);
}
ll expo(ll a, ll b, ll mod)
{
    ll res = 1;
    while (b > 0)
    {
        if (b & 1)
            res = (res * a) % mod;
        a = (a * a) % mod;
        b = b >> 1;
    }
    return res;
}
void extendgcd(ll a, ll b, ll *v)
{
    if (b == 0)
    {
        v[0] = 1;
        v[1] = 0;
        v[2] = a;
        return;
    }
    extendgcd(b, a % b, v);
    ll x = v[1];
    v[1] = v[0] - v[1] * (a / b);
    v[0] = x;
    return;
}
ll mminv(ll a, ll b)
{
    ll arr[3];
    extendgcd(a, b, arr);
    return arr[0];
}
ll mminvprime(ll a, ll b) { return expo(a, b - 2, b); }
bool revsort(ll a, ll b) { return a > b; }
ll combination(ll n, ll r, ll m, ll *fact, ll *ifact)
{
    ll val1 = fact[n];
    ll val2 = ifact[n - r];
    ll val3 = ifact[r];
    return (((val1 * val2) % m) * val3) % m;
}
void google(int t) { cout << "Case #" << t << ": "; }
vector<ll> sieve(int n)
{
    int *arr = new int[n + 1]();
    vector<ll> vect;
    for (int i = 2; i <= n; i++)
        if (arr[i] == 0)
        {
            vect.push_back(i);
            for (int j = 2 * i; j <= n; j += i)
                arr[j] = 1;
        }
    return vect;
}
ll mod_add(ll a, ll b, ll m)
{
    a = a % m;
    b = b % m;
    return (((a + b) % m) + m) % m;
}
ll mod_mul(ll a, ll b, ll m)
{
    a = a % m;
    b = b % m;
    return (((a * b) % m) + m) % m;
}
ll mod_sub(ll a, ll b, ll m)
{
    a = a % m;
    b = b % m;
    return (((a - b) % m) + m) % m;
}
ll mod_div(ll a, ll b, ll m)
{
    a = a % m;
    b = b % m;
    return (mod_mul(a, mminvprime(b, m), m) + m) % m;
}
ll phin(ll n)
{
    ll number = n;
    if (n % 2 == 0)
    {
        number /= 2;
        while (n % 2 == 0)
            n /= 2;
    }
    for (ll i = 3; i <= sqrt(n); i += 2)
    {
        if (n % i == 0)
        {
            while (n % i == 0)
                n /= i;
            number = (number / i * (i - 1));
        }
    }
    if (n > 1)
        number = (number / n * (n - 1));
    return number;
}
/*--------------------------------------------------------------------------------------------------------------------------*/

ll n;
void dfs(ll node, vector<bool> &vis, vector<vector<ll>> &adj, bool &ans, ll &start, vector<pair<ll, ll>> &final, ll par, ll &last, vector<bool> &vis2)
{

    vis[node] = true;
    vis2[node] = true;
    // for (auto it : adj[node])
    // {
    //     if (!vis[it])
    //     {
    //         mp[node].insert(it);
    //     }
    // }
    for (auto it : adj[node])
    {
        // pr(node);
        // pr(it);
        // pr(par);

        if (it == par)
            continue;
        if (vis[it] && adj[it].size() >= 4)
        {
            vector<ll> v2;
            for (auto it2 : adj[it])
            {
                if (!vis[it2])
                    v2.pb(it2);
            }
            if (v2.size() >= 2)
            {
                ans = true;
                // pr(ans);
                // pr(it);

                final.pb({v2[0], it});
                final.pb({v2[1], it});
                last = it;
            }
        }
        if (!vis[it])
        {
            dfs(it, vis, adj, ans, start, final, node, last, vis2);
        }
        if (ans)
        {
            // pr(node);
            if (vis[last])
                final.pb({node, it});

            vis[node] = false;
            return;
        }
    }
    vis[node] = false;
    // for (auto it : adj[node])
    // {
    //     if (!vis[it])
    //     {
    //         mp[node].erase(it);
    //     }
    // }
    return;
}
void solve(ll n)
{
    ll k;
    cin >> k;
    vector<vector<ll>> adj(n + 1);
    for (ll i = 0; i < k; i++)
    {
        ll u, v;
        cin >> u >> v;
        adj[u].pb(v);
        adj[v].pb(u);
    }
    vector<bool> vis2(n + 1, false);
    bool ans = false;
    vector<pll> final;
    while(1);
    // map<ll, set<ll>> mp;
    ll last = -1;
    for (ll i = 1; i <= n; i++)
    {

        if (adj[i].size() >= 4)
        {
            vector<bool> vis(n + 1, false);

            dfs(i, vis, adj, ans, i, final, 0, last, vis2);

            if (ans)
            {
                cout << "YES\n";
                cout << final.size() << "\n";
                for (auto it : final)
                {
                    cout << it.first << " " << it.second << "\n";
                }
                return;
            }
        }
    }
    cout << "NO\n";
    // vector<ll>v1(n),v2;
    // input(v1);
}
int main()
{
    fastio();
    auto start1 = high_resolution_clock::now();
    ll t;
    cin >> t;
    while (t--)
    {

        cin >> n;
        solve(n);
    }
    auto stop1 = high_resolution_clock::now();
    auto duration = duration_cast<microseconds>(stop1 - start1);
}