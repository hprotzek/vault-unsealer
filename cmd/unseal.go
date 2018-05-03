package cmd

import (
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/hprotzek/vault-unsealer/pkg/vault"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const cfgUnsealPeriod = "unseal-period"
const cfgInitRootToken = "init-root-token"
const cfgStoreRootToken = "store-root-token"
const cfgOverwriteExisting = "overwrite-existing"

type unsealCfg struct {
	unsealPeriod time.Duration
}

var unsealConfig unsealCfg

// unsealCmd represents the unseal command
var unsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		appConfig.BindPFlag(cfgUnsealPeriod, cmd.PersistentFlags().Lookup(cfgUnsealPeriod))
		appConfig.BindPFlag(cfgInitRootToken, cmd.PersistentFlags().Lookup(cfgInitRootToken))
		appConfig.BindPFlag(cfgStoreRootToken, cmd.PersistentFlags().Lookup(cfgStoreRootToken))
		appConfig.BindPFlag(cfgOverwriteExisting, cmd.PersistentFlags().Lookup(cfgOverwriteExisting))

		store, err := kvStoreForConfig(appConfig)

		if err != nil {
			logrus.Fatalf("error creating kv store: %s", err.Error())
		}

		cl, err := api.NewClient(nil)

		if err != nil {
			logrus.Fatalf("error connecting to vault: %s", err.Error())
		}

		if err != nil {
			logrus.Fatalf("error building vault config: %s", err.Error())
		}

		vaultConfig, err := vaultConfigForConfig(appConfig)

		v, err := vault.New(store, cl, vaultConfig)

		if err != nil {
			logrus.Fatalf("error creating vault helper: %s", err.Error())
		}

		for {
			func() {
				initialized, err := v.Initialized()
				if err != nil {
					logrus.Errorf("error checking if vault is initialized: %s", err.Error())
					return
				}

				if !initialized {
					if err = v.Init(); err != nil {
						logrus.Fatalf("error initialising vault: %s", err.Error())
					}
				}

				logrus.Infof("checking if vault is sealed...")
				sealed, err := v.Sealed()
				if err != nil {
					logrus.Errorf("error checking if vault is sealed: %s", err.Error())
					return
				}

				logrus.Infof("vault sealed: %t", sealed)

				// If vault is not sealed, we stop here and wait another 30 seconds
				if !sealed {
					return
				}

				if err = v.Unseal(); err != nil {
					logrus.Errorf("error unsealing vault: %s", err.Error())
					return
				}

				logrus.Infof("successfully unsealed vault")
			}()
			// wait 30 seconds before trying again
			time.Sleep(time.Second * 30)
		}
	},
}

func init() {
	unsealCmd.PersistentFlags().Duration(cfgUnsealPeriod, time.Second*30, "How often to attempt to unseal the vault instance")
	unsealCmd.PersistentFlags().String(cfgInitRootToken, "", "root token for the new vault cluster")
	unsealCmd.PersistentFlags().Bool(cfgStoreRootToken, true, "should the root token be stored in the key store")
	unsealCmd.PersistentFlags().Bool(cfgOverwriteExisting, false, "overwrite existing unseal keys and root tokens, possibly dangerous!")

	RootCmd.AddCommand(unsealCmd)
}
