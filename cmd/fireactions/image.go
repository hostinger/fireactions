package main

import (
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"github.com/spf13/cobra"
)

// newImageCmd returns the parent image command with all subcommands
func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image",
		Short:   "Manage images",
		Long:    "Manage container images - list and remove images.",
		GroupID: "image",
	}

	cmd.PersistentFlags().StringP("endpoint", "e", "127.0.0.1:8080", "Sets the Fireactions server endpoint")

	cmd.AddGroup(&cobra.Group{ID: "image", Title: "Image management commands:"})
	cmd.AddCommand(newImageListCmd())
	cmd.AddCommand(newImageRemoveCmd())

	return cmd
}

func newImageListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all images",
		RunE:    runImageListCmd,
		Args:    cobra.NoArgs,
		Aliases: []string{"ls"},
	}

	return cmd
}

func runImageListCmd(cmd *cobra.Command, _ []string) error {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	resp, err := client.ListImages(cmd.Context(), &serverv1.ListImagesRequest{})
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}

	printer.PrintText(&printableImage{resp.Images}, cmd.OutOrStdout(), nil)
	return nil
}

func newImageRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove NAME",
		Short:   "Remove an image",
		RunE:    runImageRemoveCmd,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"rm"},
	}

	return cmd
}

func runImageRemoveCmd(cmd *cobra.Command, args []string) error {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	_, err = client.RemoveImage(cmd.Context(), &serverv1.RemoveImageRequest{Name: args[0]})
	if err != nil {
		return fmt.Errorf("remove image \"%s\": %w", args[0], err)
	}

	fmt.Printf("Image \"%s\" removed\n", args[0])
	return nil
}
