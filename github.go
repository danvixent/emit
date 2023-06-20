package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

func addGithubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "github",
		Short:            "Emit a github event",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		RunE: func(cmd *cobra.Command, args []string) error {

			body := map[string]interface{}{
				"product_id": "123",
			}

			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return err
			}

			secret, err := cmd.Flags().GetString("secret")
			if err != nil {
				return err
			}

			idemKey, err := cmd.Flags().GetString("idem-key")
			if err != nil {
				return err
			}

			if url == "" || secret == "" {
				return errors.New("url and secret are required to emit github event")
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

			hmacHeader, err := generateGithubHMAC(secret, b)
			if err != nil {
				return fmt.Errorf("failed to generate hmac header new http request: %v", err)
			}

			setGithubHeaders(r, hmacHeader, idemKey)

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

func generateGithubHMAC(secret string, payload []byte) (string, error) {
	return generateHMAC(secret, payload, "hex")
}

func setGithubHeaders(r *http.Request, hmacHeader, idemKey string) {
	r.Header.Set("X-Hub-Signature-256", "sha256="+hmacHeader)
	r.Header.Set("X-Idempotency-Key", idemKey)
}
