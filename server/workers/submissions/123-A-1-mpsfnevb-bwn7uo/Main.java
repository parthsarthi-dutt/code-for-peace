import java.io.*;
import java.util.*;

public class Main {

    static final long MOD = 998244353L;

    static long modMul(long a, long b, long mod) {
        a %= mod;
        b %= mod;
        return ((a * b) % mod + mod) % mod;
    }

    static long dfs(
            int i,
            HashMap<Long, Long> next,
            HashMap<Long, Long> prev,
            long[] v2,
            HashSet<Long> s2) {

        if (i == v2.length) {
            return 1;
        }

        boolean flag = false;
        long ans = 0;
        long cur = v2[i];

        if (next.containsKey(cur) && !s2.contains(next.get(cur))) {

            long a = next.get(cur);

            if (next.containsKey(a)) {
                long b = next.get(a);
                next.put(cur, b);
                prev.put(b, cur);
                next.remove(a);
            } else {
                next.remove(cur);
                flag = true;
            }

            prev.remove(a);

            s2.remove(cur);

            ans = dfs(i + 1, next, prev, v2, s2);

            s2.add(cur);

            if (flag) {
                next.put(cur, a);
                prev.put(a, cur);
            } else {
                long b = next.get(cur);
                next.put(cur, a);
                prev.put(b, a);
                next.put(a, b);
                prev.put(a, cur);
            }

            if (prev.containsKey(cur) && !s2.contains(prev.get(cur))) {
                ans = modMul(2, ans, MOD);
            }

            return ans;
        }

        else if (prev.containsKey(cur) && !s2.contains(prev.get(cur))) {

            long a = prev.get(cur);

            if (prev.containsKey(a)) {
                long b = prev.get(a);
                prev.put(cur, b);
                next.put(b, cur);
                prev.remove(a);
            } else {
                prev.remove(cur);
                flag = true;
            }

            next.remove(a);

            s2.remove(cur);

            ans = dfs(i + 1, next, prev, v2, s2);

            s2.add(cur);

            if (flag) {
                prev.put(cur, a);
                next.put(a, cur);
            } else {
                long b = prev.get(cur);
                prev.put(cur, a);
                next.put(b, a);
                prev.put(a, b);
                next.put(a, cur);
            }

            return ans;
        }

        return ans;
    }

    static void solve(FastScanner fs) throws Exception {

        int n = fs.nextInt();
        int k = fs.nextInt();

        long[] v1 = new long[n];
        long[] v2 = new long[k];

        for (int i = 0; i < n; i++) {
            v1[i] = fs.nextLong();
        }

        for (int i = 0; i < k; i++) {
            v2[i] = fs.nextLong();
        }

        HashMap<Long, Long> prev = new HashMap<>();
        HashMap<Long, Long> next = new HashMap<>();

        HashSet<Long> s2 = new HashSet<>();

        for (long x : v2) {
            s2.add(x);
        }

        for (int i = 0; i < n; i++) {
            if (i < n - 1) {
                next.put(v1[i], v1[i + 1]);
            }

            if (i > 0) {
                prev.put(v1[i], v1[i - 1]);
            }
        }

        long ans = dfs(0, next, prev, v2, s2);
        System.out.println(ans);
    }

    public static void main(String[] args) throws Exception {

        FastScanner fs = new FastScanner(System.in);

        int t = fs.nextInt();

        while (t-- > 0) {
            solve(fs);
        }
    }

    static class FastScanner {

        private final InputStream in;
        private final byte[] buffer = new byte[1 << 16];
        private int ptr = 0, len = 0;

        FastScanner(InputStream is) {
            in = is;
        }

        private int read() throws IOException {
            if (ptr >= len) {
                len = in.read(buffer);
                ptr = 0;
                if (len <= 0) return -1;
            }
            return buffer[ptr++];
        }

        long nextLong() throws IOException {
            long sign = 1;
            long val = 0;
            int c;

            do {
                c = read();
            } while (c <= ' ');

            if (c == '-') {
                sign = -1;
                c = read();
            }

            while (c > ' ') {
                val = val * 10 + (c - '0');
                c = read();
            }

            return val * sign;
        }

        int nextInt() throws IOException {
            return (int) nextLong();
        }
    }
}