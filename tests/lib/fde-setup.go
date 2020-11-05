package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snapcore/snapd/osutil"
)

// super secure crypto
func xor13(bs []byte) []byte {
	out := make([]byte, len(bs))
	for i := range bs {
		out[i] = bs[i] ^ 0x13
	}
	return out
}

// XXX: imort struct from snapd instead?
type fdeJSON struct {
	FdeKey       []byte `json:"fde-key,omitempty"`
	FdeSealedKey []byte `json:"fde-sealed-key,omitempty"`

	VolumeName       string `json:"fde-volume-name,omitempty"`
	SourceDevicePath string `json:"fde-source-device-path,omitempty"`
}

func runFdeSetup() error {
	output, err := exec.Command("snapctl", "fde-setup-request").CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot run snapctl fde-setup-request: %v", osutil.OutputErr(output, err))
	}
	var js fdeJSON
	if err := json.Unmarshal(output, &js); err != nil {
		return err
	}

	// "unseal"
	unsealedKey := xor13(js.FdeSealedKey)
	output, err = exec.Command("snapctl", "fde-setup-result", string(unsealedKey)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot run snapctl fde-setup-result: %v", osutil.OutputErr(output, err))
	}

	return nil
}

func runFdeRevealKey() error {
	var js fdeJSON

	if err := json.NewDecoder(os.Stdin).Decode(&js); err != nil {
		return err
	}
	// "seal"
	sealedKey := xor13(js.FdeKey)
	fmt.Fprintf(os.Stdout, "%s", sealedKey)

	return nil
}

func main() {
	var err error

	switch filepath.Base(os.Args[0]) {
	case "fde-setup":
		// run as regular hook
		err = runFdeSetup()
	case "fde-reveal-key":
		// run from initrd
		err = runFdeRevealKey()
	default:
		err = fmt.Errorf("binary needs to be called as fde-setup or fde-reveal-key")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
