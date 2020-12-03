package main

import (
	"bytes"
	"encoding/hex"
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

// XXX: import struct from snapd instead?
type fdeSetupJSON struct {
	// XXX: make "op" a type: "initial-setup", "update"
	Op string `json:"op"`

	Key     []byte `json:"key,omitempty"`
	KeyName string `json:"key-name,omitempty"`

	Model map[string]string `json:"model,omitempty"`
}

func runFdeSetup() error {
	output, err := exec.Command("snapctl", "fde-setup-request").CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot run snapctl fde-setup-request: %v", osutil.OutputErr(output, err))
	}
	var js fdeSetupJSON
	if err := json.Unmarshal(output, &js); err != nil {
		return err
	}

	var fdeSetupResultStr string
	switch js.Op {
	case "features":
		// no special features supported by this hook
		fdeSetupResultStr = "[]"
	case "initial-setup":
		// "seal"
		fdeSetupResultStr = hex.EncodeToString(xor13(js.Key))
	default:
		return fmt.Errorf("unsupported op %q", js.Op)
	}
	cmd := exec.Command("snapctl", "fde-setup-result")
	cmd.Stdin = bytes.NewBufferString(fdeSetupResultStr)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot run snapctl fde-setup-result for op %q: %v", js.Op, osutil.OutputErr(output, err))
	}
	return nil
}

type fdeRevealJSON struct {
	Op string `json:"op"`

	SealedKey     []byte `json:"sealed-key"`
	SealedKeyName string `json:"sealed-key-name"`
}

func runFdeRevealKey() error {
	var js fdeRevealJSON

	if err := json.NewDecoder(os.Stdin).Decode(&js); err != nil {
		return err
	}
	switch js.Op {
	case "reveal", "lock":
		// good
	default:
		return fmt.Errorf(`only "reveal,lock" operations are supported`)
	}

	// "unseal"
	sealedKey, err := hex.DecodeString(string(js.SealedKey))
	if err != nil {
		return fmt.Errorf("cannot decode %s: %v", js.SealedKey, err)
	}
	unsealedKey := xor13(sealedKey)
	fmt.Fprintf(os.Stdout, "%s", unsealedKey)

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
