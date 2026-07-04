package cmd

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Bekican/gorenel/internal/mcp"
)

var (
	mcpCommand string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp [port]",
	Short: "Expose a local MCP (Model Context Protocol) server",
	Long: `Expose a local MCP server to the internet. Supports both existing HTTP/SSE MCP servers
and running stdio-based MCP servers using a built-in stdio-to-SSE bridge.

Examples:
  # Expose an existing HTTP/SSE MCP server running on port 8080:
  gorenel mcp 8080

  # Spawn and expose a stdio-based MCP server using Node:
  gorenel mcp --command "node my-mcp-server.js"

  # Expose a python stdio MCP server with secure key authentication:
  gorenel mcp --command "python server.py" --key-auth "my-secret-token"`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Set tunnel type to mcp
		tunnelType = "mcp"
		viper.Set("type", "mcp")

		if mcpCommand != "" {
			// Stdio bridge mode
			log.Println("Exposing stdio MCP server using built-in SSE bridge...")

			// Start local listener on random free port
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				log.Fatalf("Failed to start local bridge listener: %v", err)
			}
			addr := listener.Addr().(*net.TCPAddr)
			localPort = addr.Port
			viper.Set("port", localPort)

			// Instantiate MCP Bridge
			// Split command and arguments if any (for simple exec)
			parts := strings.Fields(mcpCommand)
			bridgeCmd := parts[0]
			var bridgeArgs []string
			if len(parts) > 1 {
				bridgeArgs = parts[1:]
			}
			bridge := mcp.NewBridge(bridgeCmd, bridgeArgs)

			// Start the bridge subprocess
			ctx := context.Background()
			if err := bridge.Start(ctx); err != nil {
				log.Fatalf("Failed to start MCP subprocess: %v", err)
			}

			// Start local HTTP server for the bridge
			httpServer := &http.Server{
				Handler: bridge,
			}
			go func() {
				if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
					log.Printf("MCP bridge HTTP server error: %v", err)
				}
			}()

			// Shut down the bridge when tunnel exits
			defer func() {
				bridge.Stop()
				httpServer.Close()
				listener.Close()
			}()

			log.Printf("Local MCP SSE bridge listening on localhost:%d", localPort)

			// Continue to run start
			runStart(cmd, args)
		} else {
			// Existing HTTP/SSE port mode
			if len(args) != 1 {
				log.Fatal("Either a port (e.g. 'gorenel mcp 8080') or a command (--command) is required.")
			}
			portVal, err := strconv.Atoi(args[0])
			if err != nil || portVal <= 0 || portVal > 65535 {
				log.Fatalf("Invalid port: %s", args[0])
			}
			localPort = portVal
			viper.Set("port", portVal)

			// Continue to run start
			runStart(cmd, args)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server address")
	mcpCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Custom subdomain")
	mcpCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key")
	mcpCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Custom domain")
	mcpCmd.Flags().StringVar(&mcpCommand, "command", "", "Stdio command to execute and bridge (e.g. 'node server.js')")
	mcpCmd.Flags().StringVar(&keyAuthToken, "key-auth", "", "Secure your MCP tunnel with a token (header X-TOKEN)")
	mcpCmd.Flags().StringVar(&authCredentials, "auth", "", "Password protect your tunnel (form: 'user:pass')")
	mcpCmd.Flags().StringArrayVar(&ipWhitelist, "ip-whitelist", []string{}, "Allowed client IP/CIDR (repeatable)")
	mcpCmd.Flags().BoolVar(&corsEnabled, "cors", true, "Enable built-in Smart CORS handling (default: true for MCP)")
	mcpCmd.Flags().StringVar(&preferRegion, "region", "", "Prefer a Fly.io region for tunnel control-plane")
	mcpCmd.Flags().BoolVar(&stableSubdomain, "stable", false, "Use stable subdomain based on project+port")

	viper.BindPFlag("server", mcpCmd.Flags().Lookup("server"))
	viper.BindPFlag("api_key", mcpCmd.Flags().Lookup("key"))
	viper.BindPFlag("domain", mcpCmd.Flags().Lookup("domain"))
	viper.BindPFlag("region", mcpCmd.Flags().Lookup("region"))
	viper.BindPFlag("stable", mcpCmd.Flags().Lookup("stable"))
}
