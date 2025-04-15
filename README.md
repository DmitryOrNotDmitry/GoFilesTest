# Go Test With Files
This is a lightweight framework for function testing. It is based on reading input and expected output data from files and comparing them with the actual execution results.
The tool is especially useful when working with systems like LeetCode, where it is not possible to run programs with custom test cases, or the provided test set is insufficient.
## Running Tests
The `ftest.go` file contains the `RunTests()` function, which is used to run the tests.
```GO
RunTests(
	folderPath string,           
	filesPattern string,
	process func(io.Reader) string, 
	isParallel bool, 
	stopOnFirstFail bool
)
```
1. `folderPath` – path to the folder containing input and output data files.
2. `filesPattern` – filename pattern for input files.
    An **input** file **must** contain the substring `input` in its name.
    The corresponding **expected output** file **must** have the same name but with the substring `input` replaced by `output`.
    Example of correct naming:
    - input file: `inputA.txt`
    - expected output file: `outputA.txt`
3. `process` – a function that reads the input data, calls the function under test, and returns the actual result as a string.
4. `isParallel` – whether to run tests in parallel.
5. `stopOnFirstFail` – whether to stop execution upon the first failed test.
## Example
Let's test a function that sums two numbers.
```go
func Sum(a, b int) int {
    return a + b
}
```
Input/output files can contain multiple test cases, separated by `@TEST SEPARATOR@`. Let's create an input file `input1.txt` with two test cases.
```txt
10 9
@TEST SEPARATOR@
15 2
```
And an output file `output1.txt` with the expected results.
```txt
19
@TEST SEPARATOR@
17
```
We then create a `Process` function that reads the input, calls the tested function, and returns the actual result as a string.
```go
func Process(readerIn io.Reader) string {
    var a, b int
    fmt.Fscan(readerIn, &a)
    fmt.Fscan(readerIn, &b)

    actual := fmt.Sprintln(Sum(a, b))
    return actual
}
```
Assuming the test files are located in the `sumTests` folder, we call `RunTests()` in the `main` function:
```go
func main() {
    ftest.RunTests("sumTests", "input*.txt", Process, true, true)
}
```
Running the program will produce the following output.
```txt
[OK] - All tests (2) successfully passed!
Total duration: 516.3µs
```
In case of an error, the output will look like this.
```txt
----------[FAIL] test 2----------
Input:
15 2
Expected:
17
Actual:
18
```
## Why `io.Reader`?
The `process` function uses `io.Reader` to easily switch the input method. In the `main` function, you can replace the `RunTests()` call with a direct call to `Process()` and read input from the console.
```go
func main() {
    fmt.Print(Process(os.Stdin))
}
```