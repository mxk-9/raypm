package progress

import (
	"fmt"
	"testing"
	"time"
)

func TestPrintingLines(t *testing.T) {
	t.Run("printing 'Hello', than 'Bye'", func(t *testing.T){
		fmt.Printf("Hello (wait 2 secs)")
		time.Sleep(time.Second * 2)
		fmt.Printf(ClearLine)
		fmt.Printf("\r")
		fmt.Printf("Bye\n")
	})
}
