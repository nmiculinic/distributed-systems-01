package main

import "github.com/spf13/cobra"

func main() {
	send := &cobra.Command{
		Use:   "send",
		Short: "sends data",
	}

	recv := &cobra.Command{
		Use:   "recv",
		Short: "recv data",
	}

	root := &cobra.Command{
		Use: "main",
	}
	root
}
