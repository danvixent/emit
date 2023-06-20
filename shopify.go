package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

func addShopifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "shopify",
		Short:            "Emit a shopify event",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		RunE: func(cmd *cobra.Command, args []string) error {

			body := map[string]interface{}{
				"product_id": "123",
			}

			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return err
			}

			idemKey, err := cmd.Flags().GetString("idem-key")
			if err != nil {
				return err
			}

			secret, err := cmd.Flags().GetString("secret")
			if err != nil {
				return err
			}

			if url == "" || secret == "" {
				return errors.New("url and secret are required to emit shopify event")
			}
			fmt.Println("url", url)
			fmt.Println("idem key", idemKey)
			fmt.Println("secret", secret)

			b, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal body: %v", err)
			}

			buf := bytes.NewBuffer(b)
			r, err := http.NewRequest(http.MethodPost, url, buf)
			if err != nil {
				return fmt.Errorf("failed to create new http request: %v", err)
			}

			hmacHeader, err := generateShopifyHMAC(secret, b)
			if err != nil {
				return fmt.Errorf("failed to generate hmac header new http request: %v", err)
			}

			setShopifyHeaders(r, hmacHeader, idemKey)

			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				return fmt.Errorf("failed to send request: %v", err)
			}

			respBuf, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to decode response body: %v", err)
			}

			fmt.Printf("response body: \n%s\n", string(respBuf))
			fmt.Printf("status code: %d", resp.StatusCode)

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	}

	return cmd
}

func generateShopifyHMAC(secret string, payload []byte) (string, error) {
	return generateHMAC(secret, payload, "base64")
}

func generateHMAC(secret string, payload []byte, encoding string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	computedMAC := mac.Sum(nil)

	switch encoding {
	case "hex":
		return hex.EncodeToString(computedMAC), nil
	case "base64":
		b := base64.StdEncoding.EncodeToString(computedMAC)
		fmt.Println("b", b)
		return b, nil
	default:
		return "", fmt.Errorf("unknown encoding type: %s", encoding)
	}
}

func setShopifyHeaders(r *http.Request, hmacHeader string, idemKey string) {
	r.Header.Set("X-Shopify-Topic", "orders/create")
	r.Header.Set("X-Idempotency-Key", idemKey)
	r.Header.Set("X-Shopify-Hmac-SHA256", hmacHeader)
	r.Header.Set("X-Shopify-Shop-Domain", "emit-test-domain")
	r.Header.Set("X-Shopify-API-Version", "2022-07")
	r.Header.Set("X-Shopify-Webhook-Id", uuid.NewString())
}
