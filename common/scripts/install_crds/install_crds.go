package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/telekom/controlplane-mono/common/pkg/test/testutil"
)

const reStr = "github.com/telekom/controlplane-mono/.*/api"

func main() {
	paths := testutil.GetCrdPathsOrDie(reStr)
	if len(paths) == 0 {
		log.Print("Nothing to apply")
	}
	for _, path := range paths {
		log.Printf("Applying CRD: %s", path)
		if _, err := os.Stat(path); err != nil {
			log.Printf("CRD file not found: %s", path)
			continue
		}
		cmd := exec.Command("kubectl", "apply", "-f", path)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print(string(out))
		fmt.Println(strings.Repeat("-", 80))

		log.Print("I'm done with my job.")

	}
}
