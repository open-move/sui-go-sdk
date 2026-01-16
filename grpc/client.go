package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient aggregates strongly-typed gRPC clients for the Sui RPC services.
type GRPCClient struct {
	endpoint string
	conn     *grpc.ClientConn

	ledgerClient                v2.LedgerServiceClient
	movePackageClient           v2.MovePackageServiceClient
	nameServiceClient           v2.NameServiceClient
	signatureVerificationClient v2.SignatureVerificationServiceClient
	stateClient                 v2.StateServiceClient
	subscriptionClient          v2.SubscriptionServiceClient
	transactionExecutionClient  v2.TransactionExecutionServiceClient
}

// NewClient dials the provided endpoint and returns a GRPCClient that wraps the generated protobuf stubs.
// Callers are responsible for closing the client via Close when they are finished with it.
func NewClient(ctx context.Context, endpoint string, opts ...Option) (*GRPCClient, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	target, serverName, secure, err := normalizeEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	dialOpts := make([]grpc.DialOption, 0, len(cfg.dialOptions)+2)
	creds, err := selectCredentials(cfg, secure, serverName)
	if err != nil {
		return nil, err
	}
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	dialOpts = append(dialOpts, cfg.dialOptions...)

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial %q: %w", endpoint, err)
	}

	c := &GRPCClient{
		conn:                        conn,
		endpoint:                    endpoint,
		ledgerClient:                v2.NewLedgerServiceClient(conn),
		movePackageClient:           v2.NewMovePackageServiceClient(conn),
		nameServiceClient:           v2.NewNameServiceClient(conn),
		signatureVerificationClient: v2.NewSignatureVerificationServiceClient(conn),
		stateClient:                 v2.NewStateServiceClient(conn),
		subscriptionClient:          v2.NewSubscriptionServiceClient(conn),
		transactionExecutionClient:  v2.NewTransactionExecutionServiceClient(conn),
	}

	return c, nil
}

// NewMainnetClient constructs a GRPCClient that targets the public Sui mainnet fullnode.
func NewMainnetClient(ctx context.Context, opts ...Option) (*GRPCClient, error) {
	return NewClient(ctx, MainnetFullnodeURL, opts...)
}

// NewTestnetClient constructs a GRPCClient that targets the public Sui testnet fullnode.
func NewTestnetClient(ctx context.Context, opts ...Option) (*GRPCClient, error) {
	return NewClient(ctx, TestnetFullnodeURL, opts...)
}

// NewDevnetClient constructs a GRPCClient that targets the public Sui devnet fullnode.
func NewDevnetClient(ctx context.Context, opts ...Option) (*GRPCClient, error) {
	return NewClient(ctx, DevnetFullnodeURL, opts...)
}

// Endpoint reports the remote endpoint the client was created for.
func (c *GRPCClient) Endpoint() string {
	if c == nil {
		return ""
	}
	return c.endpoint
}

// Conn exposes the underlying grpc.ClientConn for advanced use cases.
func (c *GRPCClient) Conn() *grpc.ClientConn {
	if c == nil {
		return nil
	}
	return c.conn
}

// Close shuts down the underlying gRPC connection.
func (c *GRPCClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// LedgerClient returns the generated LedgerService client for advanced RPC access.
func (c *GRPCClient) LedgerClient() v2.LedgerServiceClient {
	return c.ledgerClient
}

// MovePackageClient returns the generated MovePackageService client for advanced RPC access.
func (c *GRPCClient) MovePackageClient() v2.MovePackageServiceClient {
	return c.movePackageClient
}

// NameServiceClient returns the generated NameService client for advanced RPC access.
func (c *GRPCClient) NameServiceClient() v2.NameServiceClient {
	return c.nameServiceClient
}

// SignatureVerificationClient returns the generated SignatureVerificationService client for advanced RPC access.
func (c *GRPCClient) SignatureVerificationClient() v2.SignatureVerificationServiceClient {
	return c.signatureVerificationClient
}

// StateClient returns the generated StateService client for advanced RPC access.
func (c *GRPCClient) StateClient() v2.StateServiceClient {
	return c.stateClient
}

// SubscriptionClient returns the generated SubscriptionService client for advanced RPC access.
func (c *GRPCClient) SubscriptionClient() v2.SubscriptionServiceClient {
	return c.subscriptionClient
}

// TransactionExecutionClient returns the generated TransactionExecutionService client for advanced RPC access.
func (c *GRPCClient) TransactionExecutionClient() v2.TransactionExecutionServiceClient {
	return c.transactionExecutionClient
}

func selectCredentials(cfg *config, secure bool, serverName string) (credentials.TransportCredentials, error) {
	if cfg.transportCredentials != nil {
		return cfg.transportCredentials, nil
	}

	if !secure {
		return insecure.NewCredentials(), nil
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.tlsConfig != nil {
		tlsCfg = cfg.tlsConfig.Clone()
		if tlsCfg.MinVersion == 0 {
			tlsCfg.MinVersion = tls.VersionTLS12
		}
	}
	if tlsCfg.ServerName == "" {
		tlsCfg.ServerName = serverName
	}

	return credentials.NewTLS(tlsCfg), nil
}

func normalizeEndpoint(raw string) (target string, serverName string, secure bool, err error) {
	if strings.TrimSpace(raw) == "" {
		return "", "", false, errors.New("endpoint is empty")
	}

	if strings.Contains(raw, "://") {
		return normalizeURL(raw)
	}

	if strings.Contains(raw, "//") {
		return "", "", false, fmt.Errorf("unsupported endpoint format: %q", raw)
	}

	if strings.Contains(raw, ":") {
		host, port, splitErr := net.SplitHostPort(raw)
		if splitErr != nil {
			return "", "", false, fmt.Errorf("invalid endpoint %q: %w", raw, splitErr)
		}
		return net.JoinHostPort(host, port), host, false, nil
	}

	// Bare host defaults to TLS on the standard HTTPS port.
	return net.JoinHostPort(raw, "443"), raw, true, nil
}

func normalizeURL(raw string) (string, string, bool, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", false, fmt.Errorf("parse endpoint %q: %w", raw, err)
	}
	if u.Host == "" {
		return "", "", false, fmt.Errorf("endpoint %q is missing a host", raw)
	}
	if u.Path != "" && u.Path != "/" {
		return "", "", false, fmt.Errorf("endpoint %q must not contain a path", raw)
	}

	host := u.Hostname()
	port := u.Port()
	scheme := strings.ToLower(u.Scheme)

	switch scheme {
	case "https", "grpcs":
		if port == "" {
			port = "443"
		}
		return net.JoinHostPort(host, port), host, true, nil
	case "http", "grpc":
		if port == "" {
			port = "80"
		}
		return net.JoinHostPort(host, port), host, false, nil
	default:
		return "", "", false, fmt.Errorf("unsupported scheme %q in endpoint %q", u.Scheme, raw)
	}
}
