package serviceaccounts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// accountFlags holds the flags shared by the create and update commands.
type accountFlags struct {
	description string
	role        string
	labels      string
	active      bool
}

// fieldFlags maps each JSON field of a service account to the flag that sets it.
var fieldFlags = []struct{ field, flag string }{
	{"description", "description"},
	{"role", "role"},
	{"labels", "labels"},
	{"active", "enabled"},
}

// addAccountFlags adds the service account flags to the given command.
func addAccountFlags(cmd *cobra.Command, f *accountFlags) {
	cmd.Flags().StringVar(&f.description, "description", "", "the description of the service account")
	cmd.Flags().StringVar(&f.role, "role", "", "the role of the service account. It is required to create a service account")
	cmd.Flags().StringVar(&f.labels, "labels", "", `the labels of the service account in the format "keyA:valueA,keyB:valueB"`)
	cmd.Flags().BoolVar(&f.active, "enabled", true, "whether the service account is enabled")
}

// anyFlagChanged returns true if the user set any of the service account flags.
func anyFlagChanged(cmd *cobra.Command) bool {
	for _, ff := range fieldFlags {
		if cmd.Flags().Changed(ff.flag) {
			return true
		}
	}
	return false
}

// parseLabels converts "keyA:valueA,keyB" into [{"key":"keyA","value":"valueA"},{"key":"keyB"}].
// Labels are validated by Discovery Core.
func parseLabels(labels string) []map[string]string {
	parsed := []map[string]string{}
	for _, label := range strings.Split(labels, ",") {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		key, value, hasValue := strings.Cut(label, ":")
		entry := map[string]string{"key": key}
		if hasValue {
			entry["value"] = value
		}
		parsed = append(parsed, entry)
	}
	return parsed
}

// formatLabels converts the labels array of a service account into the "keyA:valueA,keyB:valueB" format.
func formatLabels(labels gjson.Result) string {
	formatted := []string{}
	for _, label := range labels.Array() {
		entry := label.Get("key").String()
		if value := label.Get("value"); value.Exists() {
			entry += ":" + value.String()
		}
		formatted = append(formatted, entry)
	}
	return strings.Join(formatted, ",")
}

// setField sets a field of the service account JSON from its string representation.
func setField(config, field, value string) (string, error) {
	switch field {
	case "labels":
		return sjson.Set(config, field, parseLabels(value))
	case "active":
		active, err := strconv.ParseBool(value)
		if err != nil {
			return "", cli.NewErrorWithCause(cli.ErrorExitCode, err, "Invalid value for active: %q", value)
		}
		return sjson.Set(config, field, active)
	default:
		return sjson.Set(config, field, value)
	}
}

// applyFlags overwrites the fields of the given service account JSON with the flags the user set.
func applyFlags(cmd *cobra.Command, config string, f accountFlags) (string, error) {
	values := map[string]string{
		"description": f.description,
		"role":        f.role,
		"labels":      f.labels,
		"active":      strconv.FormatBool(f.active),
	}

	var err error
	for _, ff := range fieldFlags {
		if cmd.Flags().Changed(ff.flag) {
			if config, err = setField(config, ff.field, values[ff.field]); err != nil {
				return "", err
			}
		}
	}
	return config, nil
}

// askAccount asks the user for every field of the service account, using the given JSON as the default values.
// If the user inputs an empty string, the value is not changed.
// If the user inputs a space, the value is set to an empty string.
func askAccount(ios *iostreams.IOStreams, config string) (string, error) {
	prompts := []struct{ label, field string }{
		{"Name", "name"},
		{"Description", "description"},
		{"Role", "role"},
		{"Labels", "labels"},
		{"Active", "active"},
	}

	for _, p := range prompts {
		current := gjson.Get(config, p.field)
		value := current.String()
		if p.field == "labels" {
			value = formatLabels(current)
		}

		input, err := ios.AskUser(fmt.Sprintf("%s [%s]: ", p.label, value))
		if err != nil {
			return "", cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not read the %s", strings.ToLower(p.label))
		}

		switch input {
		case "":
			continue
		case " ":
			input = ""
		}

		if config, err = setField(config, p.field, input); err != nil {
			return "", err
		}
	}
	return config, nil
}
