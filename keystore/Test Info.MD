# Test Info
 - `keystore_test.go` : ( `KS-test.gdbconf`, `KS-test/...` )
   Tests the keystore package and query execution on a keystore.
   **Clean-UP**:
    - Delete all `.gdbs` files except for: `0.gdbs`, `1.gdbs`, and `2.gdbs`
    - Remove all 'guest' entries created by previous tests in `0.gdbs`, `1.gdbs`, and `2.gdbs`
    - DO NOT forget to leave a new line at the end of each `.gdbs` file

 - `keystoreBench_test.go`
   Benchmarks common keystore functions and usecases.