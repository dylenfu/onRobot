package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
)

func Exec(filepath string) {
	var errStdout, errStderr error

	cmd := exec.Command(filepath)
	cmd.Env = addEnv()

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	stdout := NewCapturingPassThroughWriter(os.Stdout)
	stderr := NewCapturingPassThroughWriter(os.Stderr)

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	// waiting for execution end
	{
		go func() {
			_, errStdout = io.Copy(stdout, stdoutIn)
		}()
		go func() {
			_, errStderr = io.Copy(stderr, stderrIn)
		}()
		if err := cmd.Wait(); err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
	}

	// capture and print result
	if isError(errStderr) || isError(errStdout) {
		log.Fatalf("failed to capture stdout or stderr\n")
	}

	//outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	//log.Infof("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}

const (
	EnvWorkspace      = "PaletteWorkspace"
	EnvNodeIndexStart = "PaletteNodeIndexStart"
	EnvNodeNumber     = "PaletteNodeNumber"
	EnvNodeIndexEnd   = "PaletteNodeIndexEnd"
	EnvNetworkID      = "PaletteNetworkID"
	EnvStartRPCPort   = "PaletteStartRPCPort"
	EnvStartP2PPort   = "PaletteStartP2PPort"
	EnvLogLevel       = "PaletteLogLevel"
	EnvRPCAddress     = "PaletteRPCAddress"
)

func addEnv() []string {
	env := config.Conf.Environment
	list := make([]string, 0)
	var add = func(env string, value interface{}) {
		list = append(list, fmt.Sprintf("%s=%v", env, value))
	}

	add(EnvWorkspace, env.Workspace)
	add(EnvNodeIndexStart, env.NodeIdxStart)
	add(EnvNodeNumber, env.NodeNum)
	add(EnvNodeIndexEnd, int(env.NodeIdxStart+env.NodeNum-1))
	add(EnvNetworkID, env.NetworkID)
	add(EnvStartRPCPort, env.StartRPCPort)
	add(EnvStartP2PPort, env.StartP2PPort)
	add(EnvLogLevel, env.LogLevel)
	add(EnvRPCAddress, env.RPCAddress)

	return append(os.Environ(), list...)
}

// CapturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type CapturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// NewCapturingPassThroughWriter creates new CapturingPassThroughWriter
func NewCapturingPassThroughWriter(w io.Writer) *CapturingPassThroughWriter {
	return &CapturingPassThroughWriter{
		w: w,
	}
}

func (w *CapturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *CapturingPassThroughWriter) Bytes() []byte {
	return w.buf.Bytes()
}

func isError(err error) bool {
	if err != nil && err == os.ErrClosed {
		return true
	}
	return false
}
