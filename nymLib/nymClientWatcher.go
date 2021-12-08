package nymLib

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func StartEternityServerNymClientWatcher() {
	cmd := exec.Command("nym/target/release/nym-client", "run", "--id", "eternClient", "--gateway", "6LdVTJhRfJKsrUtnjFqE3TpEbCYs3VZoxmaoNFqRWn4x")

	// create a pipe for the output of the script
	// TODO pipe stderr too
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("\t > %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
	// 	return
	// }

	err = cmd.Wait()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
	// 	return
	// }

}
