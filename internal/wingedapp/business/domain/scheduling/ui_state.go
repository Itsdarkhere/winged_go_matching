package scheduling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"

	"github.com/google/uuid"
)

// GetUIState returns the complete UI state for a date instance.
// This is the unified endpoint for iOS client to get all UI information.
func (b *Business) GetUIState(
	ctx context.Context,
	dateInstanceID uuid.UUID,
	requestingUserID uuid.UUID,
) (*schedulingLib.UIStateResponse, error) {
	if b.dateInstanceFetcher == nil {
		return nil, errors.New("date instance fetcher not configured")
	}

	exec := b.transactor.DB()

	// Get the date instance with UI state
	di, err := b.dateInstanceFetcher.DateInstanceForUser(ctx, exec, dateInstanceID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("get date instance: %w", err)
	}

	// Build and return the full UI state response
	return buildUIStateResponseFromDI(di), nil
}

// ExecuteAction validates and executes an action on a date instance.
// This is the unified endpoint for iOS client to execute any action.
func (b *Business) ExecuteAction(
	ctx context.Context,
	dateInstanceID uuid.UUID,
	requestingUserID uuid.UUID,
	req schedulingLib.ActionRequest,
) (*schedulingLib.ActionResponse, error) {
	if b.dateInstanceFetcher == nil {
		return nil, errors.New("date instance fetcher not configured")
	}

	// Start transaction
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	// Get current date instance to determine state and role
	di, err := b.dateInstanceFetcher.DateInstanceForUser(ctx, tx, dateInstanceID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("get date instance: %w", err)
	}

	// Compute current UI state
	uiState := computeUIStateFromDI(di)

	// Validate action is allowed
	if err := schedulingLib.ValidateAction(req.Action, uiState, di.MyRole); err != nil {
		return &schedulingLib.ActionResponse{
			Success: false,
			Action:  req.Action,
			Error:   err.Error(),
		}, nil
	}

	// Marshal payload for routing
	var payloadBytes []byte
	if req.Payload != nil {
		payloadBytes, err = json.Marshal(req.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
	}

	// Route to appropriate handler
	result, err := b.routeAction(ctx, dateInstanceID, requestingUserID, req.Action, payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("route action: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// routeAction routes the action to the appropriate executor.
func (b *Business) routeAction(
	ctx context.Context,
	dateInstanceID, requestingUserID uuid.UUID,
	action string,
	payload []byte,
) (*schedulingLib.ActionResponse, error) {
	switch action {
	// Time flow actions
	case schedulingLib.ActionSuggestTimes:
		return b.handleSuggestTimes(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionConfirmTime:
		return b.handleConfirmTime(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionRejectTimes:
		return b.handleRejectTimes(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionRequestMoreTimes:
		return b.handleRequestMoreTimes(ctx, dateInstanceID, requestingUserID)

	// Venue flow actions
	case schedulingLib.ActionSelectVenue:
		return b.handleSelectVenue(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionConfirmBooking:
		return b.handleConfirmBooking(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionSuggestVenue:
		return b.handleSuggestVenue(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionRequestVenueChange:
		return b.handleRequestVenueChange(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionRespondVenueSuggestion:
		return b.handleRespondVenueSuggestion(ctx, dateInstanceID, requestingUserID, payload)

	// Modification actions
	case schedulingLib.ActionChangeTime:
		return b.handleChangeTime(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionChangePlace:
		return b.handleChangePlace(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionCancel:
		return b.handleCancel(ctx, dateInstanceID, requestingUserID, payload)

	// Feedback actions
	case schedulingLib.ActionDidMeet:
		return b.handleDidYouMeet(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionDecision:
		return b.handleDecision(ctx, dateInstanceID, requestingUserID, payload)

	// Logistics actions
	case schedulingLib.ActionArrived:
		return b.handleArrived(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionRunningLate:
		return b.handleRunningLate(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionNeedHelp:
		return b.handleNeedHelp(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionCancelNow:
		return b.handleCancelNow(ctx, dateInstanceID, requestingUserID)

	// Date type action
	case schedulingLib.ActionSelectDateType:
		return b.handleSelectDateType(ctx, dateInstanceID, requestingUserID, payload)

	// Venue confirmation sub-flow actions
	case schedulingLib.ActionConfirmVenueProceed:
		return b.handleConfirmVenueProceed(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionGoBackVenue:
		return b.handleGoBackVenue(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionAcceptProposedVenue:
		return b.handleAcceptProposedVenue(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionRejectProposedVenue:
		return b.handleRejectProposedVenue(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionProvideOwnVenue:
		return b.handleProvideOwnVenue(ctx, dateInstanceID, requestingUserID)

	// Booking reminder flow actions
	case schedulingLib.ActionSetReminder:
		return b.handleSetReminder(ctx, dateInstanceID, requestingUserID, payload)
	case schedulingLib.ActionChooseNewVenue:
		return b.handleChooseNewVenue(ctx, dateInstanceID, requestingUserID)
	case schedulingLib.ActionKeepReminder:
		return b.handleKeepReminder(ctx, dateInstanceID, requestingUserID)

	// Final confirmation action
	case schedulingLib.ActionConfirmAttendance:
		return b.handleConfirmAttendance(ctx, dateInstanceID, requestingUserID)

	// UI-only actions
	case schedulingLib.ActionExpandVenues,
		schedulingLib.ActionCollapseVenues,
		schedulingLib.ActionExpandTimeSlots,
		schedulingLib.ActionCollapseTimeSlots,
		schedulingLib.ActionDismissSheet:
		return &schedulingLib.ActionResponse{
			Success:    true,
			Action:     action,
			Message:    "UI navigation action",
			NavigateTo: action,
		}, nil

	default:
		return &schedulingLib.ActionResponse{
			Success: false,
			Action:  action,
			Error:   fmt.Sprintf("unknown action: %s", action),
		}, nil
	}
}

// computeUIStateFromDI computes the UI state from a DateInstanceUI.
func computeUIStateFromDI(di *schedulingLib.DateInstanceUI) schedulingLib.UIStateName {
	input := schedulingLib.UIStateInput{
		StatusName:            di.Status,
		HasProposals:          di.HasPendingProposals,
		AllProposalsRejected:  di.AllProposalsRejected,
		HasVenue:              di.VenueRefID != nil,
		HasDateType:           di.DateTypeCore != nil,
		IsWithinActiveWindow:  isDateActive(di),
		BookingFailed:         di.BookingStatus != nil && *di.BookingStatus == schedulingLib.BookingStatusBookingFailed,
		VenueProposalAccepted: di.VenueProposalStatus != nil && *di.VenueProposalStatus == schedulingLib.VenueProposalStatusAccepted,
		InitiatorConfirmed:    di.InitiatorConfirmedAt != nil,
		ReceiverConfirmed:     di.ReceiverConfirmedAt != nil,
	}
	return schedulingLib.ComputeUIState(input)
}

func isDateActive(di *schedulingLib.DateInstanceUI) bool {
	if di.ScheduledTimeUTC == nil {
		return false
	}
	now := time.Now()
	start := di.ScheduledTimeUTC.Add(-30 * time.Minute)
	end := di.ScheduledTimeUTC.Add(2 * time.Hour)
	return now.After(start) && now.Before(end)
}

func isPastDate(di *schedulingLib.DateInstanceUI) bool {
	if di.ScheduledTimeUTC == nil {
		return false
	}
	return time.Now().After(di.ScheduledTimeUTC.Add(2 * time.Hour))
}

// buildUIStateResponseFromDI constructs the UI state response from DateInstanceUI.
func buildUIStateResponseFromDI(di *schedulingLib.DateInstanceUI) *schedulingLib.UIStateResponse {
	uiState := computeUIStateFromDI(di)
	now := time.Now()

	response := &schedulingLib.UIStateResponse{
		DateInstanceID:    di.ID,
		MatchResultRefID:  di.MatchResultRefID,
		UIState:           uiState,
		UIStateCode:       string(uiState),
		Hint:              schedulingLib.GetHintForState(uiState, di.MyRole),
		MyRole:            di.MyRole,
		PartnerInfo:       di.PartnerInfo,
		StatusName:        di.Status,
		DecisionWindowEnd: di.DecisionWindowEnd,
		GeneratedAt:       now,
		Elements:          []schedulingLib.UIElement{},
		AvailableActions:  []schedulingLib.UIAction{},
	}

	if di.ScheduledTimeUTC != nil {
		response.ScheduledTimeUTC = di.ScheduledTimeUTC
	}
	if di.VenueRefID != nil {
		response.VenueRefID = di.VenueRefID
	}
	if di.VenueName != nil {
		response.VenueName = di.VenueName
	}

	// Build elements based on state
	response.Elements = buildElementsForState(uiState, di)

	// Build available actions
	response.AvailableActions = buildActionsForState(uiState, di.MyRole)

	// Build status line
	response.StatusLine = buildStatusLine(uiState, di.MyRole)

	// Build timer if applicable
	if shouldShowTimer(uiState) {
		response.Timer = buildTimer(di)
	}

	return response
}

func buildElementsForState(state schedulingLib.UIStateName, di *schedulingLib.DateInstanceUI) []schedulingLib.UIElement {
	elements := []schedulingLib.UIElement{}
	order := 0

	// Always add status line
	order++
	elements = append(elements, schedulingLib.UIElement{
		Type:  schedulingLib.UIElementTypeStatusLine,
		Order: order,
	})

	switch state {
	case schedulingLib.UIStateAwaitingTimeConfirmation:
		order++
		elements = append(elements, schedulingLib.UIElement{
			Type:  schedulingLib.UIElementTypeTimeSlots,
			Order: order,
		})

	case schedulingLib.UIStateSelectingVenue, schedulingLib.UIStateVenueProposedToReceiver:
		order++
		elements = append(elements, schedulingLib.UIElement{
			Type:       schedulingLib.UIElementTypeVenueOptions,
			Order:      order,
			Expandable: true,
		})

	case schedulingLib.UIStateAwaitingBooking:
		order++
		elements = append(elements, schedulingLib.UIElement{
			Type:  schedulingLib.UIElementTypeBookingStatus,
			Order: order,
		})

	case schedulingLib.UIStateDateScheduled:
		order++
		elements = append(elements, schedulingLib.UIElement{
			Type:  schedulingLib.UIElementTypeConfirmation,
			Order: order,
		})

	case schedulingLib.UIStateLogisticsPanel:
		order++
		elements = append(elements, schedulingLib.UIElement{
			Type:  schedulingLib.UIElementTypeLogisticsPanel,
			Order: order,
		})

	case schedulingLib.UIStateAwaitingFeedback:
		order++
		elements = append(elements, schedulingLib.UIElement{
			Type:  schedulingLib.UIElementTypeFeedbackForm,
			Order: order,
		})
	}

	return elements
}

func buildActionsForState(state schedulingLib.UIStateName, userRole string) []schedulingLib.UIAction {
	actions := []schedulingLib.UIAction{}

	for action, transition := range schedulingLib.ActionGuards {
		allowed := false
		for _, s := range transition.FromStates {
			if s == state {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}

		if transition.AllowedRole != "both" && transition.AllowedRole != userRole {
			continue
		}

		actions = append(actions, schedulingLib.UIAction{
			Action: action,
			Label:  actionLabel(action),
			Style:  actionStyle(action),
		})
	}

	return actions
}

func buildStatusLine(state schedulingLib.UIStateName, userRole string) *schedulingLib.UIStatusLine {
	statusLines := map[schedulingLib.UIStateName]map[string]*schedulingLib.UIStatusLine{
		schedulingLib.UIStateSyncingAvailability: {
			"initiator": {Title: "Syncing schedules...", Icon: "clock"},
			"receiver":  {Title: "Syncing schedules...", Icon: "clock"},
		},
		schedulingLib.UIStateAwaitingTimeConfirmation: {
			"initiator": {Title: "Review proposed times", Subtitle: "Select a time that works for you", Icon: "calendar"},
			"receiver":  {Title: "Waiting for confirmation", Subtitle: "Your partner is reviewing the times", Icon: "clock"},
		},
		schedulingLib.UIStateSelectingVenue: {
			"initiator": {Title: "Pick a place", Subtitle: "Choose where you'd like to meet", Icon: "map"},
			"receiver":  {Title: "Partner is choosing a place", Subtitle: "They're selecting a venue", Icon: "clock"},
		},
		schedulingLib.UIStateAwaitingBooking: {
			"initiator": {Title: "Book the venue", Subtitle: "Confirm your reservation", Icon: "phone"},
			"receiver":  {Title: "Waiting for booking", Subtitle: "Your partner is confirming the reservation", Icon: "clock"},
		},
		schedulingLib.UIStateDateScheduled: {
			"initiator": {Title: "You're all set!", Subtitle: "Your date is confirmed", Icon: "check"},
			"receiver":  {Title: "You're all set!", Subtitle: "Your date is confirmed", Icon: "check"},
		},
		schedulingLib.UIStateLogisticsPanel: {
			"initiator": {Title: "Enjoy your date!", Icon: "heart"},
			"receiver":  {Title: "Enjoy your date!", Icon: "heart"},
		},
		schedulingLib.UIStateAwaitingFeedback: {
			"initiator": {Title: "How was your date?", Icon: "star"},
			"receiver":  {Title: "How was your date?", Icon: "star"},
		},
	}

	if stateLines, ok := statusLines[state]; ok {
		if line, ok := stateLines[userRole]; ok {
			return line
		}
	}

	return &schedulingLib.UIStatusLine{
		Title: string(state),
		Icon:  "info",
	}
}

func shouldShowTimer(state schedulingLib.UIStateName) bool {
	timerStates := map[schedulingLib.UIStateName]bool{
		schedulingLib.UIStateAwaitingTimeConfirmation: true,
		schedulingLib.UIStateVenueProposedToReceiver:  true,
		schedulingLib.UIStateAwaitingConfirmation:     true,
	}
	return timerStates[state]
}

func buildTimer(di *schedulingLib.DateInstanceUI) *schedulingLib.UITimer {
	remaining := int(time.Until(di.DecisionWindowEnd).Seconds())
	expired := remaining <= 0

	style := "normal"
	if remaining < 3600 {
		style = "urgent"
	}

	return &schedulingLib.UITimer{
		Label:     "Time remaining",
		EndsAt:    di.DecisionWindowEnd,
		Style:     style,
		Expired:   expired,
		Remaining: remaining,
	}
}

func actionLabel(action string) string {
	labels := map[string]string{
		schedulingLib.ActionConfirmTime:            "Confirm Time",
		schedulingLib.ActionSuggestTimes:           "Suggest Times",
		schedulingLib.ActionRejectTimes:            "None Work For Me",
		schedulingLib.ActionRequestMoreTimes:       "Need More Options",
		schedulingLib.ActionSelectVenue:            "Select Venue",
		schedulingLib.ActionConfirmBooking:         "Confirm Booking",
		schedulingLib.ActionSuggestVenue:           "Suggest a Place",
		schedulingLib.ActionRequestVenueChange:     "Request Different Place",
		schedulingLib.ActionRespondVenueSuggestion: "Respond to Suggestion",
		schedulingLib.ActionChangeTime:             "Change Time",
		schedulingLib.ActionChangePlace:            "Change Place",
		schedulingLib.ActionCancel:                 "Cancel Date",
		schedulingLib.ActionConfirmAttendance:      "I'll Be There",
		schedulingLib.ActionArrived:                "I've Arrived",
		schedulingLib.ActionRunningLate:            "Running Late",
		schedulingLib.ActionNeedHelp:               "Need Help",
		schedulingLib.ActionCancelNow:              "Cancel Now",
		schedulingLib.ActionDidMeet:                "Submit Feedback",
		schedulingLib.ActionDecision:               "Make Decision",
	}
	if label, ok := labels[action]; ok {
		return label
	}
	return action
}

func actionStyle(action string) string {
	destructive := map[string]bool{
		schedulingLib.ActionRejectTimes: true,
		schedulingLib.ActionCancel:      true,
		schedulingLib.ActionCancelNow:   true,
	}
	if destructive[action] {
		return "destructive"
	}

	primary := map[string]bool{
		schedulingLib.ActionConfirmTime:       true,
		schedulingLib.ActionSelectVenue:       true,
		schedulingLib.ActionConfirmBooking:    true,
		schedulingLib.ActionConfirmAttendance: true,
		schedulingLib.ActionArrived:           true,
	}
	if primary[action] {
		return "primary"
	}

	return "secondary"
}

// Action handlers - delegate to existing executors

func (b *Business) handleSuggestTimes(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	var p schedulingLib.PayloadSuggestTimes
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.timeFlowExecutor.SuggestDateInstanceTimes(ctx, b.transactor.DB(), &schedulingLib.SuggestDateInstanceTimesParams{
		DateInstanceID:         dateInstanceID,
		RequestingUserID:       requestingUserID,
		ProposedScheduledTimes: p.ProposedTimes,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionSuggestTimes, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionSuggestTimes,
		Message: fmt.Sprintf("Suggested %d time(s)", result.InsertedProposalCount),
	}, nil
}

func (b *Business) handleConfirmTime(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	var p schedulingLib.PayloadConfirmTime
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.timeFlowExecutor.ConfirmDateInstanceTime(ctx, b.transactor.DB(), &schedulingLib.ConfirmDateInstanceTimeParams{
		DateInstanceID:        dateInstanceID,
		RequestingUserID:      requestingUserID,
		SelectedScheduledTime: p.SelectedTime,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionConfirmTime, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionConfirmTime,
		Message: fmt.Sprintf("Confirmed time: %s", result.ConfirmedScheduledTime.Format(time.RFC3339)),
	}, nil
}

func (b *Business) handleRejectTimes(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	var p schedulingLib.PayloadRejectTimes
	if payload != nil {
		_ = json.Unmarshal(payload, &p)
	}

	result, err := b.timeFlowExecutor.RejectDateInstanceTimes(ctx, b.transactor.DB(), &schedulingLib.RejectDateInstanceTimesParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Reason:           p.Reason,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionRejectTimes, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionRejectTimes,
		Message: fmt.Sprintf("Rejected %d proposal(s)", result.RejectedProposalCount),
	}, nil
}

func (b *Business) handleRequestMoreTimes(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	result, err := b.timeFlowExecutor.RequestMoreTimes(ctx, b.transactor.DB(), &schedulingLib.RequestMoreTimesParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionRequestMoreTimes, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionRequestMoreTimes,
		Message: result.Message,
	}, nil
}

func (b *Business) handleSelectVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	var p schedulingLib.PayloadSelectVenue
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.venueFlowExecutor.SelectVenue(ctx, b.transactor.DB(), &schedulingLib.SelectVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		VenueID:          p.VenueID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionSelectVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionSelectVenue,
		Message: fmt.Sprintf("Selected venue: %s", result.SelectedVenueName),
	}, nil
}

func (b *Business) handleConfirmBooking(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	var p schedulingLib.PayloadConfirmBooking
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.venueFlowExecutor.ConfirmBooking(ctx, b.transactor.DB(), &schedulingLib.ConfirmBookingParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		BookingStatus:    p.BookingStatus,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionConfirmBooking, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionConfirmBooking,
		Message: fmt.Sprintf("Booking status: %s", result.BookingStatus),
	}, nil
}

func (b *Business) handleSuggestVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueSuggestionExec == nil {
		return nil, errors.New("venue suggestion executor not configured")
	}

	var p schedulingLib.PayloadSuggestVenue
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.venueSuggestionExec.SuggestVenue(ctx, b.transactor.DB(), &schedulingLib.SuggestVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		VenueLink:        p.VenueLink,
		VenueName:        p.VenueName,
		VenueArea:        p.VenueArea,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionSuggestVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionSuggestVenue,
		Message: result.Message,
	}, nil
}

func (b *Business) handleRequestVenueChange(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	var p schedulingLib.PayloadRequestVenueChange
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.venueFlowExecutor.RequestVenueChange(ctx, b.transactor.DB(), &schedulingLib.RequestVenueChangeParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Reason:           p.Reason,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionRequestVenueChange, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionRequestVenueChange,
		Message: result.Message,
	}, nil
}

func (b *Business) handleRespondVenueSuggestion(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueSuggestionExec == nil {
		return nil, errors.New("venue suggestion executor not configured")
	}

	var p schedulingLib.PayloadRespondVenueSuggestion
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	actionStr := schedulingLib.VenueSuggestionShowAlternatives
	if p.Accept {
		actionStr = schedulingLib.VenueSuggestionAccepted
	}

	result, err := b.venueSuggestionExec.RespondToVenueSuggestion(ctx, b.transactor.DB(), &schedulingLib.RespondToVenueSuggestionParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Action:           actionStr,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionRespondVenueSuggestion, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionRespondVenueSuggestion,
		Message: result.Message,
	}, nil
}

func (b *Business) handleChangeTime(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	result, err := b.modificationExecutor.ChangeTime(ctx, b.transactor.DB(), &schedulingLib.ChangeTimeParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionChangeTime, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionChangeTime,
		Message: result.Message,
	}, nil
}

func (b *Business) handleChangePlace(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	result, err := b.modificationExecutor.ChangePlace(ctx, b.transactor.DB(), &schedulingLib.ChangePlaceParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionChangePlace, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionChangePlace,
		Message: result.Message,
	}, nil
}

func (b *Business) handleCancel(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	var p schedulingLib.PayloadCancel
	if payload != nil {
		_ = json.Unmarshal(payload, &p)
	}

	result, err := b.modificationExecutor.CancelDate(ctx, b.transactor.DB(), &schedulingLib.CancelDateParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Reason:           p.Reason,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionCancel, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionCancel,
		Message: result.Message,
	}, nil
}

func (b *Business) handleDidYouMeet(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.feedbackExecutor == nil {
		return nil, errors.New("feedback executor not configured")
	}

	var p schedulingLib.PayloadDidYouMeet
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.feedbackExecutor.SubmitDidYouMeet(ctx, b.transactor.DB(), &schedulingLib.SubmitDidYouMeetParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		DidMeet:          p.DidMeet,
		FeedbackText:     p.FeedbackText,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionDidMeet, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionDidMeet,
		Message: result.Message,
	}, nil
}

func (b *Business) handleDecision(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.feedbackExecutor == nil {
		return nil, errors.New("feedback executor not configured")
	}

	var p schedulingLib.PayloadDecision
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.feedbackExecutor.SubmitDecision(ctx, b.transactor.DB(), &schedulingLib.SubmitDecisionParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Decision:         p.Decision,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionDecision, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionDecision,
		Message: result.Message,
	}, nil
}

func (b *Business) handleArrived(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	result, err := b.logisticsExecutor.LogisticsArrived(ctx, b.transactor.DB(), &schedulingLib.LogisticsArrivedParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionArrived, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionArrived,
		Message: result.Message,
	}, nil
}

func (b *Business) handleRunningLate(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	var p schedulingLib.PayloadRunningLate
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.logisticsExecutor.LogisticsRunningLate(ctx, b.transactor.DB(), &schedulingLib.LogisticsRunningLateParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Minutes:          p.Minutes,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionRunningLate, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionRunningLate,
		Message: result.Message,
	}, nil
}

func (b *Business) handleNeedHelp(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	result, err := b.logisticsExecutor.LogisticsNeedHelp(ctx, b.transactor.DB(), &schedulingLib.LogisticsNeedHelpParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionNeedHelp, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionNeedHelp,
		Message: result.Message,
	}, nil
}

func (b *Business) handleCancelNow(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	result, err := b.logisticsExecutor.LogisticsCancelInWindow(ctx, b.transactor.DB(), &schedulingLib.LogisticsCancelInWindowParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionCancelNow, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionCancelNow,
		Message: result.Message,
	}, nil
}

// ============================================================================
// DATE TYPE HANDLER
// ============================================================================

func (b *Business) handleSelectDateType(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	var p schedulingLib.PayloadSelectDateType
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	result, err := b.venueFlowExecutor.SelectDateType(ctx, b.transactor.DB(), &schedulingLib.SelectDateTypeParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		DateType:         p.DateType,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionSelectDateType, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success:    true,
		Action:     schedulingLib.ActionSelectDateType,
		Message:    fmt.Sprintf("Selected date type: %s", result.DateType),
		NextAction: schedulingLib.ActionSelectVenue,
	}, nil
}

// ============================================================================
// VENUE CONFIRMATION SUB-FLOW HANDLERS
// ============================================================================

func (b *Business) handleConfirmVenueProceed(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	result, err := b.venueFlowExecutor.ConfirmVenueProceed(ctx, b.transactor.DB(), &schedulingLib.ConfirmVenueProceedParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionConfirmVenueProceed, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionConfirmVenueProceed,
		Message: result.Message,
	}, nil
}

func (b *Business) handleGoBackVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	result, err := b.venueFlowExecutor.GoBackVenue(ctx, b.transactor.DB(), &schedulingLib.GoBackVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionGoBackVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success:    true,
		Action:     schedulingLib.ActionGoBackVenue,
		Message:    result.Message,
		NavigateTo: "selecting_venue",
	}, nil
}

func (b *Business) handleAcceptProposedVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	result, err := b.venueFlowExecutor.AcceptProposedVenue(ctx, b.transactor.DB(), &schedulingLib.AcceptProposedVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionAcceptProposedVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success:    true,
		Action:     schedulingLib.ActionAcceptProposedVenue,
		Message:    result.Message,
		NextAction: schedulingLib.ActionConfirmBooking,
	}, nil
}

func (b *Business) handleRejectProposedVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	var p schedulingLib.PayloadRejectProposedVenue
	if payload != nil {
		_ = json.Unmarshal(payload, &p)
	}

	result, err := b.venueFlowExecutor.RejectProposedVenue(ctx, b.transactor.DB(), &schedulingLib.RejectProposedVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		Reason:           p.Reason,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionRejectProposedVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success:    true,
		Action:     schedulingLib.ActionRejectProposedVenue,
		Message:    result.Message,
		NavigateTo: "selecting_venue",
	}, nil
}

func (b *Business) handleProvideOwnVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	result, err := b.modificationExecutor.ProvideOwnVenue(ctx, b.transactor.DB(), &schedulingLib.ProvideOwnVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionProvideOwnVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success:    true,
		Action:     schedulingLib.ActionProvideOwnVenue,
		Message:    result.Message,
		NavigateTo: "custom_venue_input",
	}, nil
}

// ============================================================================
// BOOKING REMINDER FLOW HANDLERS
// ============================================================================

func (b *Business) handleSetReminder(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID, payload []byte) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	var p schedulingLib.PayloadSetReminder
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	remindAt, err := time.Parse(time.RFC3339, p.RemindAt)
	if err != nil {
		return nil, fmt.Errorf("invalid remind_at timestamp: %w", err)
	}

	result, err := b.modificationExecutor.SetReminder(ctx, b.transactor.DB(), &schedulingLib.SetReminderParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
		RemindAt:         remindAt,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionSetReminder, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionSetReminder,
		Message: result.Message,
	}, nil
}

func (b *Business) handleChooseNewVenue(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	result, err := b.modificationExecutor.ChooseNewVenue(ctx, b.transactor.DB(), &schedulingLib.ChooseNewVenueParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionChooseNewVenue, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success:    true,
		Action:     schedulingLib.ActionChooseNewVenue,
		Message:    result.Message,
		NavigateTo: "selecting_venue",
	}, nil
}

func (b *Business) handleKeepReminder(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	result, err := b.modificationExecutor.KeepReminder(ctx, b.transactor.DB(), &schedulingLib.KeepReminderParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionKeepReminder, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionKeepReminder,
		Message: result.Message,
	}, nil
}

// ============================================================================
// FINAL CONFIRMATION HANDLER
// ============================================================================

func (b *Business) handleConfirmAttendance(ctx context.Context, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.ActionResponse, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	result, err := b.modificationExecutor.ConfirmAttendance(ctx, b.transactor.DB(), &schedulingLib.ConfirmAttendanceParams{
		DateInstanceID:   dateInstanceID,
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		return &schedulingLib.ActionResponse{Success: false, Action: schedulingLib.ActionConfirmAttendance, Error: err.Error()}, nil
	}

	return &schedulingLib.ActionResponse{
		Success: true,
		Action:  schedulingLib.ActionConfirmAttendance,
		Message: result.Message,
	}, nil
}
