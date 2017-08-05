package tool

import (
	"os/exec"
	"sync"
	"strings"
)

func Cmd(cmd string) ([]byte, error) {
//	Vln(4, "command: ", cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	return exec.Command(head, parts...).Output()
}

func CmdArg(cmd ...string) ([]byte, error) {
//	Vln(4, "command: ", cmd)

	return exec.Command(cmd[0], cmd[1:]...).Output()
}

func Cmds(x []string) {
	wg := new(sync.WaitGroup)
	wg.Add(len(x))

	for _, cmdstr := range x {
		go cmd_wg(cmdstr, wg)
	}

	wg.Wait()
}

func cmd_wg(cmd string, wg *sync.WaitGroup) {
//	Vln(3, "command: ", cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
//		Vln(3, "err:", err)
	}
//	Vf(3, "%s", out)
	wg.Done()
}


