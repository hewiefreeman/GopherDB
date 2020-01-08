# Keystore test info
 - `keystore_test.go` : ( `KS-test.gdbconf`, `KS-test/...` )
   Tests the keystore package and query execution on a keystore.
   **Clean-UP**:
    - Delete all `.gdbs` files except for: `0.gdbs`, `1.gdbs`, and `2.gdbs`
    - Remove all 'guest###' entries and indexes created by previous tests in `0.gdbs`, `1.gdbs`, and `2.gdbs`
    - Indexing is the array at the end of each `gdbs` file. There should only be as many indexes in the array as entries in the file.
    - Set `FileOn` in `KS-test.gdbconf` to `0`

# Keystore benchmark test info
 - `keystoreBench_test.go` : ( `KS-bench.gdbconf`, `KS-bench/...` )
   Benchmarks common keystore functions and usecases.

   (Not ready for use, nor compatable with new engine)