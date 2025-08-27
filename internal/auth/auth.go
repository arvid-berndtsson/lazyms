package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/arvid-berndtsson/lazyms/internal/config"
)

// Info holds minimal identity details for status UI.
type Info struct {
	UserPrincipalName string
	TenantID          string
}

// Authenticate obtains a token using preferred method and extracts basic claims.
// It tries CLI first when preferred is "cli", otherwise Device Code.
func Authenticate(ctx context.Context, cfg config.Config) (Info, error) {
	var (
		cred azcore.TokenCredential
		err  error
	)
	useCLI := strings.EqualFold(cfg.PreferredAuth, "cli") || cfg.PreferredAuth == ""
	if useCLI {
		cred, err = azidentity.NewAzureCLICredential(nil)
		if err != nil {
			// Try to complete az login, then retry CLI cred
			if ensureAzLogin(ctx) == nil {
				cred, err = azidentity.NewAzureCLICredential(nil)
			}
			if err != nil {
				// fallback to device code
				cred, err = deviceCodeCredential(cfg)
			}
		}
	} else {
		cred, err = deviceCodeCredential(cfg)
		if err != nil {
			cred, err = azidentity.NewAzureCLICredential(nil)
		}
	}
	if err != nil {
		return Info{}, err
	}
	// Request a Graph token to get UPN claim reliably.
	tok, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil && useCLI {
		// If CLI token fetch failed, attempt az login and retry once
		if ensureAzLogin(ctx) == nil {
			if c2, e2 := azidentity.NewAzureCLICredential(nil); e2 == nil {
				cred = c2
				tok, err = cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://graph.microsoft.com/.default"}})
			}
		}
	}
	if err != nil {
		// Final fallback: device code
		if dc, e2 := deviceCodeCredential(cfg); e2 == nil {
			cred = dc
			tok, err = cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://graph.microsoft.com/.default"}})
		}
	}
	if err != nil {
		return Info{}, err
	}
	claims, _ := parseJWTClaims(tok.Token)
	info := Info{
		UserPrincipalName: firstNonEmpty(claims["upn"], claims["preferred_username"], claims["unique_name"]),
		TenantID:          str(claims["tid"]),
	}
	return info, nil
}

// ensureAzLogin checks if Azure CLI is logged in, and if not, runs `az login`.
func ensureAzLogin(ctx context.Context) error {
	if err := exec.CommandContext(ctx, "az", "account", "show", "--only-show-errors", "-o", "none").Run(); err == nil {
		return nil
	}
	cmd := exec.CommandContext(ctx, "az", "login")
	return cmd.Run()
}

func deviceCodeCredential(cfg config.Config) (azcore.TokenCredential, error) {
	opts := &azidentity.DeviceCodeCredentialOptions{
		UserPrompt: func(ctx context.Context, msg azidentity.DeviceCodeMessage) error { return nil },
	}
	if cfg.TenantID != "" {
		opts.TenantID = cfg.TenantID
	}
	if cfg.ClientID != "" {
		opts.ClientID = cfg.ClientID
	}
	return azidentity.NewDeviceCodeCredential(opts)
}

func parseJWTClaims(token string) (map[string]string, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid token")
	}
	// JWT base64url without padding
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	res := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			res[k] = s
		}
	}
	return res, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
