package scheduling

import (
	"context"

	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// transactor is an interface for handling transactions.
type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}

// availabilityGetter contains methods to get user availability.
type availabilityGetter interface {
	UserTimeBlocks(ctx context.Context, exec boil.ContextExecutor, userID string) ([]schedulingLib.TimeBlock, error)
}

// availabilitySyncer contains methods to sync user availability.
type availabilitySyncer interface {
	SyncUserAvailability(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SyncUserAvailabilityParams) (*schedulingLib.SyncUserAvailabilityResult, error)
}

// overlapFinder contains methods to find overlapping availability.
type overlapFinder interface {
	FindOverlaps(ctx context.Context, exec boil.ContextExecutor, userAID, userBID string) ([]schedulingLib.TimeBlock, error)
}

// dateInstanceFetcher contains methods for date instance queries with UI state.
type dateInstanceFetcher interface {
	DateInstanceForUser(ctx context.Context, exec boil.ContextExecutor, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.DateInstanceUI, error)
	DateInstancesForUser(ctx context.Context, exec boil.ContextExecutor, filter *schedulingLib.QueryFilterUserDateInstances) (*schedulingLib.DateInstanceUIPaginated, error)
}

// timeFlowExecutor handles Tier 3 time scheduling operations.
type timeFlowExecutor interface {
	SuggestDateInstanceTimes(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SuggestDateInstanceTimesParams) (*schedulingLib.SuggestDateInstanceTimesResult, error)
	RequestMoreTimes(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.RequestMoreTimesParams) (*schedulingLib.RequestMoreTimesResult, error)
	ConfirmDateInstanceTime(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ConfirmDateInstanceTimeParams) (*schedulingLib.ConfirmDateInstanceTimeResult, error)
	RejectDateInstanceTimes(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.RejectDateInstanceTimesParams) (*schedulingLib.RejectDateInstanceTimesResult, error)
}

// venueFlowExecutor handles Tier 4 venue selection operations.
type venueFlowExecutor interface {
	VenueOptions(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.VenueOptionsParams) (*schedulingLib.VenueOptionsResult, error)
	SelectVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SelectVenueParams) (*schedulingLib.SelectVenueResult, error)
	ConfirmBooking(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ConfirmBookingParams) (*schedulingLib.ConfirmBookingResult, error)
	RequestVenueChange(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.RequestVenueChangeParams) (*schedulingLib.RequestVenueChangeResult, error)
	// Phase 2 venue confirmation sub-flow
	SelectDateType(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SelectDateTypeParams) (*schedulingLib.SelectDateTypeResult, error)
	ConfirmVenueProceed(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ConfirmVenueProceedParams) (*schedulingLib.ConfirmVenueProceedResult, error)
	GoBackVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.GoBackVenueParams) (*schedulingLib.GoBackVenueResult, error)
	AcceptProposedVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.AcceptProposedVenueParams) (*schedulingLib.AcceptProposedVenueResult, error)
	RejectProposedVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.RejectProposedVenueParams) (*schedulingLib.RejectProposedVenueResult, error)
}

// venueSuggestionExecutor handles Tier 4 venue suggestion operations.
type venueSuggestionExecutor interface {
	SuggestVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SuggestVenueParams) (*schedulingLib.SuggestVenueResult, error)
	RespondToVenueSuggestion(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.RespondToVenueSuggestionParams) (*schedulingLib.RespondToVenueSuggestionResult, error)
}

// modificationExecutor handles Tier 5 date modification operations.
type modificationExecutor interface {
	ChangeTime(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ChangeTimeParams) (*schedulingLib.ChangeTimeResult, error)
	ChangePlace(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ChangePlaceParams) (*schedulingLib.ChangePlaceResult, error)
	CancelDate(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.CancelDateParams) (*schedulingLib.CancelDateResult, error)
	// Phase 2 booking reminder flow
	SetReminder(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SetReminderParams) (*schedulingLib.SetReminderResult, error)
	ChooseNewVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ChooseNewVenueParams) (*schedulingLib.ChooseNewVenueResult, error)
	KeepReminder(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.KeepReminderParams) (*schedulingLib.KeepReminderResult, error)
	ProvideOwnVenue(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ProvideOwnVenueParams) (*schedulingLib.ProvideOwnVenueResult, error)
	// Phase 2 final confirmation
	ConfirmAttendance(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.ConfirmAttendanceParams) (*schedulingLib.ConfirmAttendanceResult, error)
}

// feedbackExecutor handles Tier 6 post-date feedback operations.
type feedbackExecutor interface {
	PendingFeedback(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.PendingFeedbackParams) (*schedulingLib.PendingFeedbackResult, error)
	SubmitDidYouMeet(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SubmitDidYouMeetParams) (*schedulingLib.SubmitDidYouMeetResult, error)
	SubmitDecision(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.SubmitDecisionParams) (*schedulingLib.SubmitDecisionResult, error)
}

// logisticsExecutor handles Tier 7 day-of logistics operations.
type logisticsExecutor interface {
	LogisticsArrived(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.LogisticsArrivedParams) (*schedulingLib.LogisticsArrivedResult, error)
	LogisticsRunningLate(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.LogisticsRunningLateParams) (*schedulingLib.LogisticsRunningLateResult, error)
	LogisticsNeedHelp(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.LogisticsNeedHelpParams) (*schedulingLib.LogisticsNeedHelpResult, error)
	LogisticsCancelInWindow(ctx context.Context, exec boil.ContextExecutor, params *schedulingLib.LogisticsCancelInWindowParams) (*schedulingLib.LogisticsCancelInWindowResult, error)
}

// uiStateBuilder builds the complete UI state response for a date instance.
type uiStateBuilder interface {
	BuildUIState(ctx context.Context, exec boil.ContextExecutor, dateInstanceID, requestingUserID uuid.UUID) (*schedulingLib.UIStateResponse, error)
}

// actionRouter routes and executes actions on date instances.
type actionRouter interface {
	RouteAction(ctx context.Context, exec boil.ContextExecutor, dateInstanceID, requestingUserID uuid.UUID, currentUIState schedulingLib.UIStateName, userRole string, action string, payload []byte) (*schedulingLib.ActionResponse, error)
}

// actionLogger is an interface for economy actions (attend date bonus).
type actionLogger interface {
	CreateActionLog(ctx context.Context, exec boil.ContextExecutor, inserter *economy.InsertActionLog) error
}
