# Keystore test info
 - `keystore_test.go` : ( `KS-test.gdbconf`, `KS-test/...` )
   Tests the keystore package and query execution on a keystore.
   **Clean-UP**:
    - Delete all `.gdbs` files except for: `0.gdbs`, `1.gdbs`, and `2.gdbs`
    - Remove all 'guest###' entries created by previous tests in `0.gdbs`, `1.gdbs`, and `2.gdbs`
    - Set `FileOn` in `KS-test.gdbconf` to `0`
    - DO NOT forget to leave a new line at the end of each `.gdbs` file

# Keystore benchmark test info
 - `keystoreBench_test.go` : ( `KS-bench.gdbconf`, `KS-bench/...` )
   Benchmarks common keystore functions and usecases.

   Not ready for use, nor compatable with new engine!