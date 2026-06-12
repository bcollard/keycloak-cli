// Package cliutil holds small helpers shared across cobra command packages.
package cliutil

import "github.com/spf13/pflag"

// AliasNormalizer returns a pflag normalize function that maps alias flag names
// to a canonical one, so synonyms resolve to the same flag. For example, with
// {"user": "username"} both --user and --username drive the flag registered as
// "username". Apply it with cmd.Flags().SetNormalizeFunc(AliasNormalizer(...))
// after the flags are defined.
//
// This is preferred over registering two flags pointing at the same variable:
// for repeatable (StringArray) flags that approach silently drops values when
// both names are passed, whereas normalization makes them genuinely one flag.
func AliasNormalizer(aliases map[string]string) func(*pflag.FlagSet, string) pflag.NormalizedName {
	return func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		if canonical, ok := aliases[name]; ok {
			name = canonical
		}
		return pflag.NormalizedName(name)
	}
}
