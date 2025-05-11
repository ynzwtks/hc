[日本語版の `README.md`](./README.ja.md)

# hc
hc is a CLI tool to support local testing for the AtCoder Heuristic Contest. It features a simple single-binary setup and interactive installation to minimize initial setup tasks, supporting everything from test execution to log recording.

<br>

### Screencast
![Screencast](https://github.com/ynzwtks/hc/assets/73768325/667d9765-4cfa-4793-9f54-aabfbe0d8649)

<br>


## Features
- Parallel execution of tests on the local machine
- Generation and switching of test sets
- Display of progress during test execution
- Filtering execution based on input parameters of test cases
- Log integration with the local version of the official standings tool
- Parallel execution via Cloud Run jobs (requires separate configuration)

<br>

## How to Install
Download the latest executable file (hc) from the [release](https://github.com/ynzwtks/hc/releases) and place it in any folder included in your PATH.  
Alternatively, install using the following command:

```shell
go install github.com/ynzwtks/hc@latest
```

<br>
The file structure is as follows. Files other than the CLI itself are automatically generated when the CLI is executed.

| Element  | File |
| ------------- | ------------- |
| CLI  | hc       |
| Definition File   | contest.toml   |
| Logs    |  {BaseDir}/logs/* |
| Test Case Definitions | {BaseDir}/test/* |

<br>

## How to use
Refer to the help for each command for details.

```
Usage:
  hc [command]

Available Commands:
  config      Configure settings
  jobs        Execute cloud run jobs
  log         Display results
  run         Run the test set
  web         Display standings or the histogram of parameters
```

<br>

---
## Quick Start

### 1. Preparation

Before setting up, prepare and confirm the following:
- Preparation of execution programs  
  - Tools provided by the official source (gen, vis, or tester)
  - Program to be judged
  
- Confirmation of contest evaluation method  
  - Scoring method (maximize or minimize the score)
  - Whether it is an interactive problem

<br>

### 2. Environment Setup

Initialize the config. Initialization will create contest.toml in the current directory.
```shell
hc config new
```
Set up interactively.
```shell
hc config setup
```
Add a test set (100 test cases named set1 from seed0).
```shell
hc config add test -s set1 -c 100
```

Add a system test (assuming seeds and result.csv are published by the official source).
```shell
hc config add system -n ahc031
```

List the test sets.
```shell
hc config list
```

Switch the default test set.
```shell
hc config switch -s {SetName}
```

<br>

### 3. Test Execution

Execute a specific test case. The standard output result during test case execution is output to out.txt in the current directory.
```shell
hc run -t {TestNumber}
```

Execute the test set with specified parameter conditions.
```shell
hc run -t {TestNumber} -f "N<10 && D<10"
```
Execute the test set and record to log (logs are not recorded without the "-w" option).
```shell
hc run -w "test"
```

Display only the execution results without showing the progress.
```shell
hc run -q
```

<br>

### 4. Check Execution Results

Check the summary of results.
```shell
hc log
```

Check the results of each test case.
```shell
hc log {LogNumber}
```

Check the results of each test case with specified test set and parameter conditions.
```shell
hc log {LogNumber} -s {SetName} -f "N < 20"
```

Sort and display the differences of each test case in execution results.
```shell
hc log diff {LogNumber} {LogNumber} -o asc
```

Launch the official standings tool.
```shell
hc web standings
```

<br>
<br>

---

## Tips

### 1. Run CLI from a different location than the folder containing contest.toml
The CLI reads contest.toml from the current directory, but setting the environment variable "CONTEST_CONFIG_FILE" allows the CLI to be run from any location.

``` sh
export CONTEST_CONFIG_FILE=/Users/xxxxx/vsshare/ahc/ahc031/contest.toml
```
<br>

### 2. Check standard output results during test case execution
If you specify a test case with the "-t" option in the run command, the result will be output to "out.txt" in the current directory.  
Otherwise, the result will be output to "xxxx_o.txt" under "test/{SetName/".

 <br>
 
### 3. Add system test cases to the test set
If seeds.txt is published on the official site, you can download the data and add the test set by executing the following command.

```shell
    hc config add system -n {ContestName}
    hc run -s system
```

If result.csv and input.csv for the official standings tool are published, they will be downloaded and placed automatically.  
In that case, the best score and rank in result.csv will be displayed when executing the test case.

``` shell
$hc run -t 1 -s system
No=0001 Score=15744  Best=8596 Rank=17/1062
Parameter=[1000 40 37] Seed=1464601681064286668
```

<br>

### 4. Retrieve information about which test case is being executed within the evaluation program
By referring to the environment variable "INPUT_FILE", the evaluation program can reference the filename of the currently executing test case.

<br>

### 5. Execute test cases with filtering
The first line parameters of the test cases are set with the config setup command.  
For example, if the first line of the test case refers to "W D N", input it as follows with space separation.

  ```
  $hc config setup
         :
  ? Enter the input header's fields(Optional):  W D N
  ```
<br>
You can filter by adding the "-f" option to the run and log commands, such as "-f N<10".  
<br>
  
Refer to [govaluate](https://github.com/Knetic/govaluate) for operators that can be used for filtering.
    
<br>

### 6. Filter with custom parameters
The procedure for adding parameters A and B, for example, is as follows.

  1. Place the file "ex.dat" with the parameter values for each test case in the directory where the test cases are stored (test/{SetName}).
     
     - Define multiple parameters with space separation.
     - Values must be within the range of Float64.
    
     ```
     1 2
     1 1
     2 2
     :
     ```
        
  2. Define the parameter names in the "ExFields" section of the test set definition in contest.toml.
   
  ```toml
  [{SetName}]
  SetName = '{SetName}'
  TestDataPath = 'test/{SetName}/'
  TestDataNum = 3000
  ExFields = 'A B'  
  IsSystemTest = false
  ```
    
<br>

### 7. Display only test execution results
Execute the run command with the "-q" flag. (Results are displayed in the order of geometric mean, arithmetic mean, number of test cases, and number of errors)

```shell
$hc run -q
10512 20835 30 0
```
      
<br>

### 8. Set specific environment variables during CLI execution
Add the definitions to the env section of contest.toml with the keys and values paired as shown below.

```toml
[env]
Keys = ['DEBUG']
Values = ['true']
```
    
<br>

### 9. Execute tests in parallel using Google Cloud Run Jobs
  -  To be added in the future.

<br>

## Change Log

### 2025-05-11
- Changed directory structure for test cases
  - Input files: Now placed and referenced in `./test/${setname}/in/`
  - Output files: Now output to `./test/${setname}/out/`
  - Removed `_o` suffix from output filenames to match input filenames
