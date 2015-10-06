import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// joinArgs joins strings and arrays of strings together into one array.
func joinArgs(args ...interface{}) []string {
	var final []string
	for _, a := range args {
		switch a.(type) {
		case string:
			final = append(final, a.(string))
		case []string:
			final = append(final, a.([]string)...)
		default:
			final = append(final, fmt.Sprint(a))
		}
	}
	return final
}

// system runs cmd, and prints the stderr output to ew, if ew is not nil.
func system(cmd *exec.Cmd, ew io.Writer) error {
	if ew != nil {
		return cmd.Run()
	}

	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	rd := bufio.NewReader(out)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Fprintln(ew, line)
	}

	return cmd.Wait()
}

// in performs a function in a directory, and then returns to the
// previous directory.
func in(dir string, f func() error) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(r.Directory)
	if err != nil {
		return err
	}
	defer func() {
		cerr = os.Chdir(cwd)
		if err == nil {
			err = cerr
		}
	}()
	err = f()
	return
}

