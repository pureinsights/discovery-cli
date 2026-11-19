package tokens

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// tokenFlags holds the flags shared by the create and update commands.
type tokenFlags struct {
	scopes       string
	addScopes    string
	removeScopes string
	ttl          string
	active       bool
}

// fieldFlags maps each JSON field of a token to the flag that sets it. Applied in order.
var fieldFlags = []struct{ field, flag string }{
	{"scopes", "scopes"},
	{"addScopes", "add-scopes"},
	{"removeScopes", "remove-scopes"},
	{"ttl", "ttl"},
	{"active", "enabled"},
}

// addTokenFlags adds the token flags to the given command. TTL is only set on creation or renewal.
func addTokenFlags(cmd *cobra.Command, f *tokenFlags, withTTL bool) {
	cmd.Flags().StringVar(&f.scopes, "scopes", "", "the comma-separated scopes for the token")
	if withTTL {
		cmd.Flags().StringVar(&f.ttl, "ttl", "", "the time-to-live for the token")
	}
	cmd.Flags().BoolVar(&f.active, "enabled", true, "whether the token is enabled")
}

// addScopeEditFlags adds the flags that add or remove scopes from the current scopes of a token.
func addScopeEditFlags(cmd *cobra.Command, f *tokenFlags) {
	cmd.Flags().StringVar(&f.addScopes, "add-scopes", "", "the comma-separated scopes to add to the token")
	cmd.Flags().StringVar(&f.removeScopes, "remove-scopes", "", "the comma-separated scopes to remove from the token")
}

// splitScopes splits the comma-separated scopes
func splitScopes(value string) []string {
	scopes := []string{}
	for _, scope := range strings.Split(value, ",") {
		if scope = strings.TrimSpace(scope); scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

// addServiceAccountFlag adds the required --service-account flag (name or UUID).
func addServiceAccountFlag(cmd *cobra.Command, serviceAccount *string) {
	cmd.Flags().StringVar(serviceAccount, "service-account", "", "the name or UUID of the service account that owns the token")
	_ = cmd.MarkFlagRequired("service-account")
}

// serviceAccountId resolves the service account by name or UUID to the UUID used in the token paths.
func serviceAccountId(d cli.Discovery, coreUrl, coreKey, serviceAccount string) (string, error) {
	id, err := cli.GetEntityId(d, discoveryPackage.NewCore(coreUrl, coreKey).ServiceAccounts(), serviceAccount)
	if err != nil {
		return "", cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not find service account %q", serviceAccount)
	}
	return id.String(), nil
}

// anyFlagChanged returns true if the user set any of the token flags.
func anyFlagChanged(cmd *cobra.Command) bool {
	for _, ff := range fieldFlags {
		if cmd.Flags().Changed(ff.flag) {
			return true
		}
	}
	return false
}

// setField sets a field of the token JSON from its string representation.
func setField(config, field, value string) (string, error) {
	switch field {
	case "active":
		active, err := strconv.ParseBool(value)
		if err != nil {
			return "", cli.NewErrorWithCause(cli.ErrorExitCode, err, "Invalid value for active: %q", value)
		}
		return sjson.Set(config, field, active)
	case "scopes":
		return sjson.Set(config, field, splitScopes(value))
	case "addScopes", "removeScopes":
		scopes := []string{}
		for _, scope := range gjson.Get(config, "scopes").Array() {
			scopes = append(scopes, scope.String())
		}
		for _, scope := range splitScopes(value) {
			if field == "addScopes" {
				if !slices.Contains(scopes, scope) {
					scopes = append(scopes, scope)
				}
			} else {
				scopes = slices.DeleteFunc(scopes, func(s string) bool { return s == scope })
			}
		}
		return sjson.Set(config, "scopes", scopes)
	default:
		return sjson.Set(config, field, value)
	}
}

// applyFlags overwrites the fields of the given token JSON with the flags the user set.
func applyFlags(cmd *cobra.Command, config string, f tokenFlags) (string, error) {
	values := map[string]string{
		"scopes":       f.scopes,
		"addScopes":    f.addScopes,
		"removeScopes": f.removeScopes,
		"ttl":          f.ttl,
		"active":       strconv.FormatBool(f.active),
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

// askToken asks the user for every field of the token, using the given JSON as the default values.
// If user inputs empty, value is not changed.
// If user inputs is a space, value is set to an empty string.
func askToken(ios *iostreams.IOStreams, config string, withTTL bool) (string, error) {
	prompts := []struct{ label, field string }{
		{"Name", "name"},
		{"Scopes", "scopes"},
	}
	if withTTL {
		prompts = append(prompts, struct{ label, field string }{"TTL", "ttl"})
	}
	prompts = append(prompts, struct{ label, field string }{"Active", "active"})

	for _, p := range prompts {
		current := gjson.Get(config, p.field)
		value := current.String()

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
