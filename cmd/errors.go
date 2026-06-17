package cmd

import (
	"github.com/Arubacloud/acloud-cli/internal/errs"
)

func apiErrFromV2(err error) error {
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")
	return errs.FromV2(err, debug)
}

func fmtAPIError(statusCode int, title, detail *string) error {
	return errs.FmtAPIError(statusCode, title, detail)
}
