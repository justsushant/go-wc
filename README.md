# Word Count

This excerise deals with a command-line program that implements Unix wc like functionality.

This exercise has been solved in a TDD fashion. Please refer to the execise [here](https://one2n.io/go-bootcamp/go-projects/a-game-of-pig).

## Features
This program supports the following options:
 - -l show only line count
 - -w show only word count 
 - -c show only character count

By default (absence of any options), it shows all three counts.

If more than one file is present in arguments, it shows the counts for all those files, with a total row added at the end.

If no file arguments are present, it reads from STDIN.

## Usage

1. Run the below command to build the binary. It has been saved in the bin directory.
```
make build
```

2. For using the program, arguments and options can be passed as usual. Some examples are as follows:

- For counting on a single file (for all three counts):
    ```
    ./bin/go-wc file.txt
    ```
- For counting on multiple files (word and character count):
    ```
    ./bin/go-wc file.txt file2.txt -w -c
    ```
- For counting on STDIN (only line count):
    ```
    ./bin/go-wc -l
    ```