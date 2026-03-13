package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/c0tton-fluff/sentinelone-mcp-server/client"
	"github.com/mark3labs/mcp-go/mcp"
)

var dvQueryTool = mcp.NewTool("s1_dv_query",
	mcp.WithDescription(`Run a Deep Visibility query. Returns queryId when complete. Example query: ProcessName Contains "python"`),
	mcp.WithString("query",
		mcp.Required(),
		mcp.Description(`Deep Visibility query (e.g., ProcessName Contains "python", SrcIP = "10.0.0.1")`),
	),
	mcp.WithString("fromDate",
		mcp.Required(),
		mcp.Description("Start date in ISO format (e.g., 2024-01-01T00:00:00Z)"),
	),
	mcp.WithString("toDate",
		mcp.Required(),
		mcp.Description("End date in ISO format (e.g., 2024-01-02T00:00:00Z)"),
	),
	mcp.WithArray("siteIds",
		mcp.Description("Filter by site IDs"),
		mcp.Items(map[string]any{"type": "string"}),
	),
	mcp.WithArray("groupIds",
		mcp.Description("Filter by group IDs"),
		mcp.Items(map[string]any{"type": "string"}),
	),
	mcp.WithArray("accountIds",
		mcp.Description("Filter by account IDs"),
		mcp.Items(map[string]any{"type": "string"}),
	),
)

var dvGetEventsTool = mcp.NewTool("s1_dv_get_events",
	mcp.WithDescription("Get events from a completed Deep Visibility query"),
	mcp.WithString("queryId",
		mcp.Required(),
		mcp.Description("Query ID returned from s1_dv_query"),
	),
	mcp.WithNumber("limit",
		mcp.Description("Max results (default 50, max 100)"),
	),
	mcp.WithString("cursor",
		mcp.Description("Pagination cursor"),
	),
)

func summarizeEvent(e map[string]any) string {
	timeStr := "unknown"
	if d := getStr(e, "eventTime"); d != "" {
		timeStr = formatTimeAgo(d)
	}
	eventType := fallback(getStr(e, "eventType"), "Unknown")
	process := fallback(getStr(e, "processName"), "N/A")
	agent := fallback(getStr(e, "agentName"), "Unknown")

	var details string
	srcIP := getStr(e, "srcIp")
	dstIP := getStr(e, "dstIp")
	dstPort := getStr(e, "dstPort")
	if dstPort == "" {
		dstPort = "?"
	}

	if srcIP != "" && dstIP != "" {
		details += fmt.Sprintf(" | %s -> %s:%s", srcIP, dstIP, dstPort)
	} else if dstIP != "" {
		details += fmt.Sprintf(" -> %s:%s", dstIP, dstPort)
	}
	if fp := getStr(e, "filePath"); fp != "" {
		details += " | " + truncatePath(fp, 60)
	}
	if dns := getStr(e, "dnsRequest"); dns != "" {
		details += " | DNS: " + dns
	}
	if user := getStr(e, "user"); user != "" {
		details += " | User: " + user
	}

	return fmt.Sprintf("- %s | %s | %s | %s%s", eventType, agent, process, timeStr, details)
}

func handleDVQuery(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fromDate, err := req.RequireString("fromDate")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	toDate, err := req.RequireString("toDate")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	siteIDs := req.GetStringSlice("siteIds", nil)
	groupIDs := req.GetStringSlice("groupIds", nil)
	accountIDs := req.GetStringSlice("accountIds", nil)

	queryID, err := client.CreateDVQuery(query, fromDate, toDate, siteIDs, groupIDs, accountIDs)
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("Error running Deep Visibility query: %v", err),
		), nil
	}

	// Poll for completion
	var status *client.DVStatus
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		status, err = client.GetDVQueryStatus(queryID)
		if err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("Error running Deep Visibility query: %v", err),
			), nil
		}
		if status.Status != "RUNNING" {
			break
		}
	}

	if status.Status == "FAILED" {
		return mcp.NewToolResultError(
			fmt.Sprintf("Deep Visibility query failed: %s", fallback(status.ResponseError, "Unknown error")),
		), nil
	}

	if status.Status == "RUNNING" {
		return mcp.NewToolResultText(
			fmt.Sprintf("Query still running after 30 seconds. Use s1_dv_get_events with queryId: %s to retrieve results later.", queryID),
		), nil
	}

	result := map[string]string{
		"queryId": queryID,
		"status":  status.Status,
		"message": "Query completed. Use s1_dv_get_events to retrieve results.",
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(b)), nil
}

func handleDVGetEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryID, err := req.RequireString("queryId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	limit := int(req.GetFloat("limit", 50))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	cursor := req.GetString("cursor", "")

	// Check query status first
	status, err := client.GetDVQueryStatus(queryID)
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("Error getting Deep Visibility events: %v", err),
		), nil
	}

	switch status.Status {
	case "RUNNING":
		return mcp.NewToolResultText(
			fmt.Sprintf("Query %s is still running (%d%% complete). Please wait and try again.",
				queryID, status.ProgressStatus),
		), nil
	case "FAILED":
		return mcp.NewToolResultError(
			fmt.Sprintf("Query %s failed: %s", queryID, fallback(status.ResponseError, "Unknown error")),
		), nil
	case "CANCELED":
		return mcp.NewToolResultError(
			fmt.Sprintf("Query %s was canceled", queryID),
		), nil
	}

	result, err := client.GetDVEvents(queryID, limit, cursor)
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("Error getting Deep Visibility events: %v", err),
		), nil
	}

	if len(result.Data) == 0 {
		return mcp.NewToolResultText("No events found for this query."), nil
	}

	lines := make([]string, len(result.Data))
	for i, e := range result.Data {
		lines[i] = summarizeEvent(e)
	}

	text := fmt.Sprintf("Found %d event(s):\n\n%s", len(result.Data), strings.Join(lines, "\n"))
	if result.Pagination != nil && result.Pagination.NextCursor != "" {
		text += fmt.Sprintf("\n\n[More results available - use cursor: %s]", result.Pagination.NextCursor)
	}

	return mcp.NewToolResultText(text), nil
}
