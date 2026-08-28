package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
)

// AllTools returns all MCP tool definitions.
func AllTools() []ToolDef {
	return []ToolDef{
		listThreatsTool,
		getThreatTool,
		mitigateThreatTool,
		setAnalystVerdictTool,
		setIncidentStatusTool,
		listAgentsTool,
		getAgentTool,
		isolateAgentTool,
		reconnectAgentTool,
		listAlertsTool,
		setAlertVerdictTool,
		setAlertStatusTool,
		hashReputationTool,
		dvQueryTool,
		dvGetEventsTool,
		investigateThreatTool,
		listExclusionsTool,
		createExclusionTool,
		deleteExclusionTool,
		createSTARRuleTool,
		listApplicationsTool,
	}
}

// DispatchTool routes a tool call to the appropriate handler.
// A panic in any handler is contained here so it surfaces as a tool error
// instead of killing the server mid-session.
func DispatchTool(ctx context.Context, name string, args json.RawMessage) (result ToolResult) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic in tool %s: %v\n%s", name, r, debug.Stack())
			result = toolError(fmt.Sprintf("internal error in %s: %v", name, r))
		}
	}()
	switch name {
	case "s1_list_threats":
		return handleListThreats(ctx, args)
	case "s1_get_threat":
		return handleGetThreat(ctx, args)
	case "s1_mitigate_threat":
		return handleMitigateThreat(ctx, args)
	case "s1_set_analyst_verdict":
		return handleSetAnalystVerdict(ctx, args)
	case "s1_set_incident_status":
		return handleSetIncidentStatus(ctx, args)
	case "s1_list_agents":
		return handleListAgents(ctx, args)
	case "s1_get_agent":
		return handleGetAgent(ctx, args)
	case "s1_isolate_agent":
		return handleIsolateAgent(ctx, args)
	case "s1_reconnect_agent":
		return handleReconnectAgent(ctx, args)
	case "s1_list_alerts":
		return handleListAlerts(ctx, args)
	case "s1_set_alert_verdict":
		return handleSetAlertVerdict(ctx, args)
	case "s1_set_alert_status":
		return handleSetAlertStatus(ctx, args)
	case "s1_hash_reputation":
		return handleHashReputation(ctx, args)
	case "s1_dv_query":
		return handleDVQuery(ctx, args)
	case "s1_dv_get_events":
		return handleDVGetEvents(ctx, args)
	case "s1_investigate_threat":
		return handleInvestigateThreat(ctx, args)
	case "s1_list_exclusions":
		return handleListExclusions(ctx, args)
	case "s1_create_exclusion":
		return handleCreateExclusion(ctx, args)
	case "s1_delete_exclusion":
		return handleDeleteExclusion(ctx, args)
	case "s1_create_star_rule":
		return handleCreateSTARRule(ctx, args)
	case "s1_list_applications":
		return handleListApplications(ctx, args)
	default:
		return toolError(fmt.Sprintf("unknown tool: %s", name))
	}
}
