//go:build search
// +build search

package client

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	zerr "zotregistry.dev/zot/errors"
	zcommon "zotregistry.dev/zot/pkg/common"
)

const (
	maxRetries = 20
)

func NewCveForImageCommand(searchService SearchService) *cobra.Command {
	var (
		searchedCVEID   string
		cveListSortFlag = CVEListSortFlag(SortBySeverity)
	)

	cveForImageCmd := &cobra.Command{
		Use:   "list [repo:tag]|[repo@digest]",
		Short: "List CVEs by REPO:TAG or REPO@DIGEST",
		Long:  `List CVEs by REPO:TAG or REPO@DIGEST`,
		Args:  OneImageWithRefArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			searchConfig, err := GetSearchConfigFromFlags(cmd, searchService)
			if err != nil {
				return err
			}

			err = CheckExtEndPointQuery(searchConfig, CVEListForImageQuery())
			if err != nil {
				return fmt.Errorf("%w: '%s'", err, CVEListForImageQuery().Name)
			}

			image := args[0]

			return SearchCVEForImageGQL(searchConfig, image, searchedCVEID)
		},
	}

	cveForImageCmd.Flags().StringVar(&searchedCVEID, SearchedCVEID, "", "Search for a specific CVE by name/id")
	cveForImageCmd.Flags().Var(&cveListSortFlag, SortByFlag,
		fmt.Sprintf("Options for sorting the output: [%s]", CVEListSortOptionsStr()))

	return cveForImageCmd
}

func NewImagesByCVEIDCommand(searchService SearchService) *cobra.Command {
	var (
		repo              string
		imageListSortFlag = ImageListSortFlag(SortByAlphabeticAsc)
	)

	imagesByCVEIDCmd := &cobra.Command{
		Use:   "affected [cveId]",
		Short: "List images affected by a CVE",
		Long:  `List images affected by a CVE`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if !strings.HasPrefix(args[0], "CVE") {
				return fmt.Errorf("%w: expected a cveid 'CVE-...' got '%s'", zerr.ErrInvalidCLIParameter, args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			searchConfig, err := GetSearchConfigFromFlags(cmd, searchService)
			if err != nil {
				return err
			}

			err = CheckExtEndPointQuery(searchConfig, ImageListForCVEQuery())
			if err != nil {
				return fmt.Errorf("%w: '%s'", err, ImageListForCVEQuery().Name)
			}

			searchedCVEID := args[0]

			return SearchImagesByCVEIDGQL(searchConfig, repo, searchedCVEID)
		},
	}

	imagesByCVEIDCmd.Flags().StringVar(&repo, "repo", "", "Search for a specific CVE by name/id")
	imagesByCVEIDCmd.Flags().Var(&imageListSortFlag, SortByFlag,
		fmt.Sprintf("Options for sorting the output: [%s]", ImageListSortOptionsStr()))

	return imagesByCVEIDCmd
}

func NewFixedTagsCommand(searchService SearchService) *cobra.Command {
	imageListSortFlag := ImageListSortFlag(SortByAlphabeticAsc)

	fixedTagsCmd := &cobra.Command{
		Use:   "fixed [repo] [cveId]",
		Short: "List tags where a CVE is fixed",
		Long:  `List tags where a CVE is fixed`,
		Args: func(cmd *cobra.Command, args []string) error {
			const argCount = 2

			if err := cobra.ExactArgs(argCount)(cmd, args); err != nil {
				return err
			}

			if !zcommon.CheckIsCorrectRepoNameFormat(args[0]) {
				return fmt.Errorf("%w: expected a valid repo name for first argument '%s'", zerr.ErrInvalidCLIParameter, args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			searchConfig, err := GetSearchConfigFromFlags(cmd, searchService)
			if err != nil {
				return err
			}

			err = CheckExtEndPointQuery(searchConfig, ImageListWithCVEFixedQuery())
			if err != nil {
				return fmt.Errorf("%w: '%s'", err, ImageListWithCVEFixedQuery().Name)
			}

			repo := args[0]
			searchedCVEID := args[1]

			return SearchFixedTagsGQL(searchConfig, repo, searchedCVEID)
		},
	}

	fixedTagsCmd.Flags().Var(&imageListSortFlag, SortByFlag,
		fmt.Sprintf("Options for sorting the output: [%s]", ImageListSortOptionsStr()))

	return fixedTagsCmd
}
