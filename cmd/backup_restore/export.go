package backupRestore

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewExportCommand(d cli.Discovery) *cobra.Command {
	var file string
	export := &cobra.Command{
		Use:   "export [subcommands]",
		Short: "Export all of Discovery Ingestion's entities",
		Long:  fmt.Sprintf(commands.LongConfig, "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key")).BackupRestore()
			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key")).BackupRestore()
			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key")).BackupRestore()

			clients := []cli.BackupRestoreClientEntry{
				{ Name: "core", Client: coreClient },
				{ Name: "ingestion", Client: ingestionClient },
				{ Name: "queryflow", Client: queryflowClient },
			}

			printer := cli.GetObjectPrinter(d.Config().GetString("output"))
			return d.ExportEntitiesFromClients(clients, file, printer)
		},
		Args: cobra.NoArgs,
	}

	export.Flags().StringVarP(&file, "file", "f", "", "the file in which the information of the entities is going to be saved")
	return export
}

// func NewExportCommand(d cli.Discovery) *cobra.Command {
// 	var file string
// 	export := &cobra.Command{
// 		Use:   "export [subcommands]",
// 		Short: "Export all of Discovery Ingestion's entities",
// 		Long:  fmt.Sprintf(commands.LongConfig, "Ingestion"),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			profile, err := cmd.Flags().GetString("profile")
// 			if err != nil {
// 				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
// 			}
//
// 			vpr := d.Config()
//
// 			err = commands.CheckCredentials(d, profile, "Core", "core_url")
// 			if err != nil {
// 				return err
// 			}
//
// 			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key")).BackupRestore()
// 			coreBytes, coreFileName, coreAcknowledged, coreErr := d.ExportEntities(coreClient)
//
// 			coreFileName = "core-" + coreFileName
//
// 			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key")).BackupRestore()
// 			ingestionBytes, ingestionFileName, ingestionAcknowledged, ingestionErr := d.ExportEntities(ingestionClient)
// 			ingestionFileName = "ingestion-" + ingestionFileName
//
// 			if file == "" {
// 				file = "discovery.zip"
// 			}
//
// 			zipFile, err := os.OpenFile(
// 				file,
// 				os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
// 				0o644,
// 			)
//
// 			zipWriter := zip.NewWriter(zipFile)
//
// 			files := []struct {
// 				fileName     string
// 				bytes        *[]byte
// 				acknowledged *gjson.Result
// 				err          *error
// 			}{
// 				{
// 					fileName:     coreFileName,
// 					bytes:        &coreBytes,
// 					acknowledged: &coreAcknowledged,
// 					err:          &coreErr,
// 				},
// 				{
// 					fileName:     ingestionFileName,
// 					bytes:        &ingestionBytes,
// 					acknowledged: &ingestionAcknowledged,
// 					err:          &ingestionErr,
// 				},
// 			}
//
// 			for _, file := range files {
// 				h := &zip.FileHeader{
// 					Name:     file.fileName, // path inside outer zip, e.g. "folder/a.zip"
// 					Method:   zip.Store,     // inner zip is already compressed; avoid recompressing
// 					Modified: time.Now(),    // optional, set a timestamp
// 				}
//
// 				fw, err := zipWriter.CreateHeader(h)
// 				if err != nil {
// 					*file.acknowledged = gjson.Parse(`{"acknowledged":false}`)
// 					*file.err = err
// 				}
//
// 				if _, err := fw.Write(*file.bytes); err != nil {
// 					*file.acknowledged = gjson.Parse(`{"acknowledged":false}`)
// 					*file.err = err
// 				}
// 			}
//
// 			zipWriter.Close()
// 			zipFile.Close()
//
// 			jsonResults := "{}"
// 			if coreErr != nil {
// 				acknowledgedString, _ := sjson.SetRaw(`{"acknowledged":false}`, "error", coreErr.Error())
// 				coreAcknowledged = gjson.Parse(acknowledgedString)
// 			}
// 			jsonResults, _ = sjson.SetRaw(jsonResults, "core", coreAcknowledged.Raw)
//
// 			if ingestionErr != nil {
// 				acknowledgedString, _ := sjson.SetRaw(`{"acknowledged":false}`, "error", ingestionErr.Error())
// 				ingestionAcknowledged = gjson.Parse(acknowledgedString)
// 			}
// 			jsonResults, _ = sjson.SetRaw(jsonResults, "ingestion", ingestionAcknowledged.Raw)
//
// 			printer := cli.GetObjectPrinter(d.Config().GetString("output"))
// 			if printer == nil {
// 				jsonPrinter := cli.JsonObjectPrinter(false)
// 				return jsonPrinter(*d.IOStreams(), gjson.Parse(jsonResults))
// 			} else {
// 				return printer(*d.IOStreams(), gjson.Parse(jsonResults))
// 			}
// 		},
// 		Args: cobra.NoArgs,
// 	}
//
// 	export.Flags().StringVarP(&file, "file", "f", "", "the file in which the information of the entities is going to be saved")
// 	return export
// }
