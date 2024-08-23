-c option in wc (from wc man page) returns bytes and golang len returns bytes 

issue with relative paths- suppose you're in a directory, you can't use paths like "../testdata/cmd_test/file1.txt" in tests (with os.DirFS(".")). this works with wc


// wc_test
file4.txt isn't working with tabs \t

// cmd_test
add permisson error test case


PATH BUG

possible sol:
    - this solution may result in different FS for test files and executable
    -get FS on the executable path
        https://stackoverflow.com/questions/18537257/how-to-get-the-directory-of-the-currently-running-file
    -get relative path from executable and file name to access the file


- Jpath := Join(working directory/os.Executable path, argument path)
- path := filePath.Dir(Jpath)
- os.DirFS(path)