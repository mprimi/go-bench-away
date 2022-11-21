package client

import (
	"fmt"
)

func (c *Client) logDebug(format string, args ...interface{}) {
	if c.options.verbose {
		fmt.Printf("[debug] client: "+format+"\n", args...)
	}
}

func (c *Client) logWarn(format string, args ...interface{}) {
	fmt.Printf("[warning] client: "+format+"\n", args...)
}
