package cmd

import (
	"fmt"

	"github.com/kubefirst/git-helper/internal/sync"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	syncWebhookOpts *sync.WebhookOptions = &sync.WebhookOptions{}

	allowedGitProviders []string = []string{"github", "gitlab"}
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize a git resource based on local parameters",
	Long:  `Synchronize a git resource based on local parameters`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called")
	},
}

// syncWebhookCmd represents the sync webhook command
var syncWebhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage a target repository/project webhook",
	Long:  `Manage a target repository/project webhook`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync webhook called")
	},
}

// syncWebhookCreateCmd represents the sync webhook create command
var syncWebhookCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a target repository/project webhook",
	Long:  `Create a target repository/project webhook`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync webhook create called")
	},
}

// syncWebhookDeleteCmd represents the sync webhook delete command
var syncWebhookDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a target repository/project webhook",
	Long:  `Delete a target repository/project webhook`,
	Run: func(cmd *cobra.Command, args []string) {
		err := sync.DeleteWebhook(*syncWebhookOpts)
		if err != nil {
			log.Fatalf("error running command: %s", err)
		}
	},
}

// syncNgrokAtlantisWebhook represents the sync webhook delete command
var syncNgrokAtlantisWebhookCmd = &cobra.Command{
	Use:   "ngrok-atlantis",
	Short: "Create a webhook based on an ngrok tunnel for Atlantis",
	Long:  `"Create a webhook based on an ngrok tunnel for Atlantis"`,
	Run: func(cmd *cobra.Command, args []string) {
		err := sync.SynchronizeAtlantisWebhook(*syncWebhookOpts)
		if err != nil {
			log.Fatalf("error running command: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncWebhookCmd)
	syncWebhookCmd.AddCommand(syncWebhookCreateCmd)
	syncWebhookCmd.AddCommand(syncWebhookDeleteCmd)
	syncWebhookCmd.AddCommand(syncNgrokAtlantisWebhookCmd)

	// Required flags
	var attach []*cobra.Command
	attach = append(attach, syncWebhookCreateCmd, syncWebhookDeleteCmd, syncNgrokAtlantisWebhookCmd)

	for _, command := range attach {
		command.Flags().StringVar(&syncWebhookOpts.Owner, "owner", syncWebhookOpts.Owner, "Owner - organization or primary group")
		err := command.MarkFlagRequired("owner")
		if err != nil {
			log.Fatal(err)
		}
		command.Flags().StringVar(&syncWebhookOpts.Provider, "provider", syncWebhookOpts.Provider, fmt.Sprintf("Provider - one of %s (required)", allowedGitProviders))
		err = command.MarkFlagRequired("provider")
		if err != nil {
			log.Fatal(err)
		}
		command.Flags().StringVar(&syncWebhookOpts.Repository, "repository", syncWebhookOpts.Repository, "Repository or project (required)")
		err = command.MarkFlagRequired("repository")
		if err != nil {
			log.Fatal(err)
		}
		command.Flags().StringVar(&syncWebhookOpts.Url, "url", syncWebhookOpts.Url, "URL endpoint to provide to webhook (required)")
		command.Flags().StringVar(&syncWebhookOpts.OldUrl, "old-url", syncWebhookOpts.OldUrl, "If replacing a webhook, the URL used by the existing (old) webhook")

		// Secret options
		command.Flags().BoolVar(&syncWebhookOpts.UseSecret, "use-secret", false, "Retrieve values from Kubernetes Secret")
		command.Flags().StringVar(&syncWebhookOpts.SecretName, "secret-name", syncWebhookOpts.SecretName, "Secret name (required if using --use-secret)")
		command.Flags().StringVar(&syncWebhookOpts.SecretNamespace, "namespace", syncWebhookOpts.SecretNamespace, "Namespace  (required if using --use-secret)")
		command.Flags().StringVar(&syncWebhookOpts.SecretValues, "secret-values", syncWebhookOpts.SecretValues, "Secret value keys to parse from secret (required if using --use-secret)")

		// Other options
		command.Flags().StringVar(&syncWebhookOpts.Token, "token", syncWebhookOpts.Token, "Secret token to provide to webhook")
		command.Flags().BoolVar(&syncWebhookOpts.Cleanup, "cleanup", false, "Remove tokens but don't add new ones")

		command.Flags().BoolVar(&syncWebhookOpts.KubeInClusterConfig, "use-kubeconfig-in-cluster", true, "kube config type - in-cluster (default), set to false to use local")
		command.Flags().BoolVar(&syncWebhookOpts.Restart, "restart", false, "If provided, trigger ngrok restart via ConfigMap edit")
	}
}
