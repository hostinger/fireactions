package workflowrun

import (
	"fmt"
	"os"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func Stats() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats ORGANISATION",
		Short:   "View workflow run statistics",
		Aliases: []string{"stat"},
		RunE:    runStatsCommand,
		Args:    cobra.ExactArgs(1),
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().DurationP("since", "S", 0, "Filter workflow runs by date since duration (e.g. 1h30m)")
	cmd.Flags().StringP("start", "s", "", "Filter workflow runs by date (RFC3339)")
	cmd.Flags().StringP("end", "e", "", "Filter workflow runs by date (RFC3339)")
	cmd.Flags().StringP("repositories", "r", "", "Comma separated list of repositories to filter by")
	cmd.Flags().StringP("sort", "o", "TOTAL", "Sort workflow runs by \"TOTAL\", \"TOTAL_DURATION\", \"SUCCEEDED\", \"CANCELLED\", \"FAILED\", \"AVERAGE_DURATION\", \"SUCCESS_RATE\", \"FAILURE_RATE\"")
	cmd.Flags().IntP("limit", "l", 100, "Limit the number of workflow runs")
	cmd.Flags().BoolP("asc", "a", false, "Sort workflow runs in ascending order")
	cmd.Flags().BoolP("dsc", "d", false, "Sort workflow runs in descending order")
	cmd.Flags().String("server-url", os.Getenv("FIREACTIONS_SERVER_URL"), "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")

	return cmd
}

func runStatsCommand(cmd *cobra.Command, args []string) error {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return err
	}

	repositories, err := cmd.Flags().GetString("repositories")
	if err != nil {
		return err
	}

	start, err := cmd.Flags().GetString("start")
	if err != nil {
		return err
	}

	var startTime time.Time
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return err
		}
	} else {
		startTime = time.Now().Add(-time.Hour * 24 * 7)
	}

	end, err := cmd.Flags().GetString("end")
	if err != nil {
		return err
	}

	var endTime time.Time
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return err
		}
	} else {
		endTime = time.Now()
	}

	since, err := cmd.Flags().GetDuration("since")
	if err != nil {
		return err
	}

	if since != 0 {
		startTime = time.Now().Add(-since)
	}

	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return err
	}

	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return err
	}

	sortOrder := "desc"
	if cmd.Flags().Changed("asc") {
		sortOrder = "asc"
	} else if cmd.Flags().Changed("dsc") {
		sortOrder = "desc"
	}

	client := fireactions.NewClient(fireactions.WithEndpoint(serverURL))

	workflowRunStats, _, err := client.GetWorkflowRunStats(cmd.Context(), args[0], &fireactions.WorkflowRunStatsQuery{
		Repositories: repositories, Start: startTime, End: endTime, Sort: sort, SortOrder: sortOrder, Limit: limit,
	})
	if err != nil {
		return err
	}

	if len(workflowRunStats.Stats) == 0 {
		cmd.Printf("No workflow runs found from %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
		return nil
	}

	tw := tablewriter.NewWriter(cmd.OutOrStdout())
	tw.SetColumnSeparator("")
	tw.SetRowSeparator("")
	tw.SetCenterSeparator("")
	tw.SetHeaderLine(false)
	tw.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: true})
	tw.SetNoWhiteSpace(true)
	tw.SetTablePadding("  ")
	if repositories != "" {
		tw.SetCaption(true, fmt.Sprintf("Workflow run statistics from %s to %s for top %d repositories (filter: %s), sorted by %s %s",
			startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), limit, repositories, sort, sortOrder))
	} else {
		tw.SetCaption(true, fmt.Sprintf("Workflow run statistics from %s to %s for top %d repositories, sorted by %s %s",
			startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), limit, sort, sortOrder))
	}
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})
	tw.SetHeader([]string{"Repository", "Total", "Total Duration", "Average Duration", "Succeeded", "Cancelled", "Failed", "Success Rate", "Failure Rate"})
	for _, stat := range workflowRunStats.Stats {
		tw.Append([]string{stat.Repository, fmt.Sprintf("%d", stat.Total), stat.TotalDuration.String(), fmt.Sprintf("%.2fm", stat.GetAverageDuration().Minutes()), fmt.Sprintf("%d", stat.Succeeded), fmt.Sprintf("%d", stat.Cancelled), fmt.Sprintf("%d", stat.Failed), fmt.Sprintf("%.2f%%", stat.GetSuccessRatio()), fmt.Sprintf("%.2f%%", stat.GetFailureRatio())})
	}

	tw.Append([]string{"", "", "", "", "", "", "", "", ""})
	tw.Append([]string{"TOTALS", fmt.Sprintf("%d", workflowRunStats.GetTotal()), workflowRunStats.GetTotalDuration().String(), fmt.Sprintf("%.2fm", workflowRunStats.GetAverageDuration().Minutes()), fmt.Sprintf("%d", workflowRunStats.GetSucceeded()),
		fmt.Sprintf("%d", workflowRunStats.GetCancelled()), fmt.Sprintf("%d", workflowRunStats.GetFailed()), fmt.Sprintf("%.2f%%", workflowRunStats.GetSuccessRatio()), fmt.Sprintf("%.2f%%", workflowRunStats.GetFailureRatio())})
	tw.Render()

	return nil
}
