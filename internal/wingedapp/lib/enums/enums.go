// Package enums provides type-safe string enums replacing the category/category_type tables.
// This eliminates join queries and provides compile-time validation.
package enums

// UserType represents user role classification.
type UserType string

const (
	UserTypeAdmin       UserType = "Admin"
	UserTypeRegularUser UserType = "Regular User"
)

func (e UserType) String() string { return string(e) }
func (e UserType) Valid() bool {
	switch e {
	case UserTypeAdmin, UserTypeRegularUser:
		return true
	}
	return false
}

// GeneralAIContext represents global AI context types.
type GeneralAIContext string

const (
	GeneralAIContextYourAgent GeneralAIContext = "Your Agent"
)

func (e GeneralAIContext) String() string { return string(e) }
func (e GeneralAIContext) Valid() bool {
	switch e {
	case GeneralAIContextYourAgent:
		return true
	}
	return false
}

// AIConvo represents AI conversation types.
type AIConvo string

const (
	AIConvoAI        AIConvo = "AI"
	AIConvoYourAgent AIConvo = "Your Agent"
)

func (e AIConvo) String() string { return string(e) }
func (e AIConvo) Valid() bool {
	switch e {
	case AIConvoAI, AIConvoYourAgent:
		return true
	}
	return false
}

// UserInviteCode represents invite code sources.
type UserInviteCode string

const (
	UserInviteCodeReferral UserInviteCode = "Referral"
	UserInviteCodeEvent    UserInviteCode = "Event"
)

func (e UserInviteCode) String() string { return string(e) }
func (e UserInviteCode) Valid() bool {
	switch e {
	case UserInviteCodeReferral, UserInviteCodeEvent:
		return true
	}
	return false
}

// WingsEconomyEarn represents earning action types.
type WingsEconomyEarn string

const (
	WingsEconomyEarnDailyCheckIn WingsEconomyEarn = "Daily Check-In"
)

func (e WingsEconomyEarn) String() string { return string(e) }
func (e WingsEconomyEarn) Valid() bool {
	switch e {
	case WingsEconomyEarnDailyCheckIn:
		return true
	}
	return false
}

// WingsEconomySpend represents spending action types.
type WingsEconomySpend string

const (
	WingsEconomySpendSendMessage WingsEconomySpend = "Send Message"
)

func (e WingsEconomySpend) String() string { return string(e) }
func (e WingsEconomySpend) Valid() bool {
	switch e {
	case WingsEconomySpendSendMessage:
		return true
	}
	return false
}

// WingsEconomySubscriptionPlan represents subscription plan types.
type WingsEconomySubscriptionPlan string

const (
	WingsEconomySubscriptionPlanWingedPlus WingsEconomySubscriptionPlan = "Winged+"
	WingsEconomySubscriptionPlanWingedX    WingsEconomySubscriptionPlan = "WingedX"
	WingsEconomySubscriptionPlanTopUp      WingsEconomySubscriptionPlan = "Top Up"
)

func (e WingsEconomySubscriptionPlan) String() string { return string(e) }
func (e WingsEconomySubscriptionPlan) Valid() bool {
	switch e {
	case WingsEconomySubscriptionPlanWingedPlus, WingsEconomySubscriptionPlanWingedX, WingsEconomySubscriptionPlanTopUp:
		return true
	}
	return false
}

// WingsEconomyActionLog represents economy transaction types.
type WingsEconomyActionLog string

const (
	WingsEconomyActionLogDailyCheckIn          WingsEconomyActionLog = "Daily Check-In"
	WingsEconomyActionLogSendMessage           WingsEconomyActionLog = "Send Message"
	WingsEconomyActionLogWingedXWeeklyPayment  WingsEconomyActionLog = "WingedX - Weekly Payment"
	WingsEconomyActionLogWingedXMonthlyPayment WingsEconomyActionLog = "WingedX - Monthly Payment"
	WingsEconomyActionLogWingedPlusWeekly      WingsEconomyActionLog = "Winged+ - Weekly Payment"
	WingsEconomyActionLogWingedPlusMonthly     WingsEconomyActionLog = "Winged+ - Monthly Payment"
	WingsEconomyActionLogWingedPlus3Month      WingsEconomyActionLog = "Winged+ - 3 Month Payment"
	WingsEconomyActionLogWingedPlus6Month      WingsEconomyActionLog = "Winged+ - 6 Month Payment"
	WingsEconomyActionLogTopUpMini             WingsEconomyActionLog = "Top Up - Mini"
	WingsEconomyActionLogTopUpBoost            WingsEconomyActionLog = "Top Up - Boost"
	WingsEconomyActionLogTopUpPremium          WingsEconomyActionLog = "Top Up - Premium"
)

func (e WingsEconomyActionLog) String() string { return string(e) }
func (e WingsEconomyActionLog) Valid() bool {
	switch e {
	case WingsEconomyActionLogDailyCheckIn, WingsEconomyActionLogSendMessage,
		WingsEconomyActionLogWingedXWeeklyPayment, WingsEconomyActionLogWingedXMonthlyPayment,
		WingsEconomyActionLogWingedPlusWeekly, WingsEconomyActionLogWingedPlusMonthly,
		WingsEconomyActionLogWingedPlus3Month, WingsEconomyActionLogWingedPlus6Month,
		WingsEconomyActionLogTopUpMini, WingsEconomyActionLogTopUpBoost, WingsEconomyActionLogTopUpPremium:
		return true
	}
	return false
}

// Sexuality represents sexual orientation options.
type Sexuality string

const (
	SexualityPreferNotToSay Sexuality = "Prefer not to say"
	SexualityStraight       Sexuality = "Straight"
	SexualityGay            Sexuality = "Gay"
	SexualityLesbian        Sexuality = "Lesbian"
	SexualityBisexual       Sexuality = "Bisexual"
	SexualityAsexual        Sexuality = "Asexual"
	SexualityQuestioning    Sexuality = "Questioning"
	SexualityOther          Sexuality = "Other"
)

func (e Sexuality) String() string { return string(e) }
func (e Sexuality) Valid() bool {
	switch e {
	case SexualityPreferNotToSay, SexualityStraight, SexualityGay, SexualityLesbian,
		SexualityBisexual, SexualityAsexual, SexualityQuestioning, SexualityOther:
		return true
	}
	return false
}

// MatchStatus represents match lifecycle (Active/Expired).
type MatchStatus string

const (
	MatchStatusActive  MatchStatus = "Active"
	MatchStatusExpired MatchStatus = "Expired"
)

func (e MatchStatus) String() string { return string(e) }
func (e MatchStatus) Valid() bool {
	switch e {
	case MatchStatusActive, MatchStatusExpired:
		return true
	}
	return false
}

// DateType represents user-facing date type display labels.
type DateType string

const (
	DateTypeCasualDinner DateType = "Casual dinner"
	DateTypeGlassOfWine  DateType = "Glass of wine"
	DateTypeArtExhibit   DateType = "Art exhibit"
	DateTypeCoffee       DateType = "Coffee"
	DateTypeCocktails    DateType = "Cocktails"
	DateTypeWalk         DateType = "Walk"
)

func (e DateType) String() string { return string(e) }
func (e DateType) Valid() bool {
	switch e {
	case DateTypeCasualDinner, DateTypeGlassOfWine, DateTypeArtExhibit,
		DateTypeCoffee, DateTypeCocktails, DateTypeWalk:
		return true
	}
	return false
}

// MatchUserAction represents user actions on matches.
type MatchUserAction string

const (
	MatchUserActionPending  MatchUserAction = "Pending"
	MatchUserActionProposed MatchUserAction = "Proposed"
	MatchUserActionPassed   MatchUserAction = "Passed"
)

func (e MatchUserAction) String() string { return string(e) }
func (e MatchUserAction) Valid() bool {
	switch e {
	case MatchUserActionPending, MatchUserActionProposed, MatchUserActionPassed:
		return true
	}
	return false
}

// DateTypeCore represents canonical date types with associated durations.
type DateTypeCore string

const (
	DateTypeCoreCoffee   DateTypeCore = "coffee"   // 60 minutes
	DateTypeCoreDrinks   DateTypeCore = "drinks"   // 90 minutes
	DateTypeCoreMeal     DateTypeCore = "meal"     // 120 minutes
	DateTypeCoreWalk     DateTypeCore = "walk"     // 75 minutes
	DateTypeCoreActivity DateTypeCore = "activity" // 120 minutes
)

func (e DateTypeCore) String() string { return string(e) }
func (e DateTypeCore) Valid() bool {
	switch e {
	case DateTypeCoreCoffee, DateTypeCoreDrinks, DateTypeCoreMeal, DateTypeCoreWalk, DateTypeCoreActivity:
		return true
	}
	return false
}

// DurationMinutes returns the default duration in minutes for this date type.
func (e DateTypeCore) DurationMinutes() int {
	switch e {
	case DateTypeCoreCoffee:
		return 60
	case DateTypeCoreDrinks:
		return 90
	case DateTypeCoreMeal:
		return 120
	case DateTypeCoreWalk:
		return 75
	case DateTypeCoreActivity:
		return 120
	}
	return 60 // default
}

// DietaryRestriction represents dietary preference options.
type DietaryRestriction string

const (
	DietaryRestrictionVegetarian    DietaryRestriction = "Vegetarian"
	DietaryRestrictionVegan         DietaryRestriction = "Vegan"
	DietaryRestrictionDairyFree     DietaryRestriction = "Dairy-Free"
	DietaryRestrictionGlutenFree    DietaryRestriction = "Gluten-Free"
	DietaryRestrictionAlcoholFree   DietaryRestriction = "Alcohol-Free"
	DietaryRestrictionHalal         DietaryRestriction = "Halal"
	DietaryRestrictionKosher        DietaryRestriction = "Kosher"
	DietaryRestrictionNoRestriction DietaryRestriction = "No Restrictions"
)

func (e DietaryRestriction) String() string { return string(e) }
func (e DietaryRestriction) Valid() bool {
	switch e {
	case DietaryRestrictionVegetarian, DietaryRestrictionVegan, DietaryRestrictionDairyFree,
		DietaryRestrictionGlutenFree, DietaryRestrictionAlcoholFree, DietaryRestrictionHalal,
		DietaryRestrictionKosher, DietaryRestrictionNoRestriction:
		return true
	}
	return false
}

// DateInstanceStatus represents date instance state machine states.
type DateInstanceStatus string

const (
	DateInstanceStatusProposed    DateInstanceStatus = "Proposed"
	DateInstanceStatusTimeChosen  DateInstanceStatus = "Time Chosen"
	DateInstanceStatusVenueChosen DateInstanceStatus = "Venue Chosen"
	DateInstanceStatusDateSet     DateInstanceStatus = "Date Set"
	DateInstanceStatusCompleted   DateInstanceStatus = "Completed"
	DateInstanceStatusCancelled   DateInstanceStatus = "Cancelled"
	DateInstanceStatusExpired     DateInstanceStatus = "Expired"
	DateInstanceStatusNoShow      DateInstanceStatus = "No Show"
)

func (e DateInstanceStatus) String() string { return string(e) }
func (e DateInstanceStatus) Valid() bool {
	switch e {
	case DateInstanceStatusProposed, DateInstanceStatusTimeChosen, DateInstanceStatusVenueChosen,
		DateInstanceStatusDateSet, DateInstanceStatusCompleted, DateInstanceStatusCancelled,
		DateInstanceStatusExpired, DateInstanceStatusNoShow:
		return true
	}
	return false
}

// BookingStatus represents venue booking state.
type BookingStatus string

const (
	BookingStatusUnknown         BookingStatus = "Unknown"
	BookingStatusBooked          BookingStatus = "Booked"
	BookingStatusNoBookingNeeded BookingStatus = "No Booking Needed"
	BookingStatusBookingFailed   BookingStatus = "Booking Failed"
)

func (e BookingStatus) String() string { return string(e) }
func (e BookingStatus) Valid() bool {
	switch e {
	case BookingStatusUnknown, BookingStatusBooked, BookingStatusNoBookingNeeded, BookingStatusBookingFailed:
		return true
	}
	return false
}

// FeedbackStatus represents post-date feedback state.
type FeedbackStatus string

const (
	FeedbackStatusPending          FeedbackStatus = "Pending"
	FeedbackStatusSubmitted        FeedbackStatus = "Submitted"
	FeedbackStatusAutoClosed       FeedbackStatus = "Auto Closed"
	FeedbackStatusSubmittedByAgent FeedbackStatus = "Submitted By Agent"
)

func (e FeedbackStatus) String() string { return string(e) }
func (e FeedbackStatus) Valid() bool {
	switch e {
	case FeedbackStatusPending, FeedbackStatusSubmitted, FeedbackStatusAutoClosed, FeedbackStatusSubmittedByAgent:
		return true
	}
	return false
}

// DateDecision represents post-date match continuation decision.
type DateDecision string

const (
	DateDecisionScheduleSecondDate DateDecision = "Schedule Second Date"
	DateDecisionCloseConnection    DateDecision = "Close Connection"
)

func (e DateDecision) String() string { return string(e) }
func (e DateDecision) Valid() bool {
	switch e {
	case DateDecisionScheduleSecondDate, DateDecisionCloseConnection:
		return true
	}
	return false
}

// DidYouMeet represents post-date verification options.
type DidYouMeet string

const (
	DidYouMeetYes            DidYouMeet = "Yes"
	DidYouMeetNo             DidYouMeet = "No"
	DidYouMeetPreferNotToSay DidYouMeet = "Prefer Not To Say"
)

func (e DidYouMeet) String() string { return string(e) }
func (e DidYouMeet) Valid() bool {
	switch e {
	case DidYouMeetYes, DidYouMeetNo, DidYouMeetPreferNotToSay:
		return true
	}
	return false
}

// SchedulingCardType represents UI card types in the scheduling flow.
type SchedulingCardType string

const (
	SchedulingCardTypeDietaryRestrictions SchedulingCardType = "Dietary Restrictions"
	SchedulingCardTypeCalendarConnect     SchedulingCardType = "Calendar Connect"
	SchedulingCardTypeTimeProposal        SchedulingCardType = "Time Proposal"
	SchedulingCardTypeTimeConfirmation    SchedulingCardType = "Time Confirmation"
	SchedulingCardTypeVenueProposal       SchedulingCardType = "Venue Proposal"
	SchedulingCardTypeBookingConfirmation SchedulingCardType = "Booking Confirmation"
	SchedulingCardTypeDateBooked          SchedulingCardType = "Date Booked"
	SchedulingCardTypeVenueRevision       SchedulingCardType = "Venue Revision"
	SchedulingCardTypeFeedbackRequest     SchedulingCardType = "Feedback Request"
	SchedulingCardTypePredateReminder     SchedulingCardType = "Predate Reminder"
)

func (e SchedulingCardType) String() string { return string(e) }
func (e SchedulingCardType) Valid() bool {
	switch e {
	case SchedulingCardTypeDietaryRestrictions, SchedulingCardTypeCalendarConnect,
		SchedulingCardTypeTimeProposal, SchedulingCardTypeTimeConfirmation,
		SchedulingCardTypeVenueProposal, SchedulingCardTypeBookingConfirmation,
		SchedulingCardTypeDateBooked, SchedulingCardTypeVenueRevision,
		SchedulingCardTypeFeedbackRequest, SchedulingCardTypePredateReminder:
		return true
	}
	return false
}

// SchedulingCardState represents card lifecycle states.
type SchedulingCardState string

const (
	SchedulingCardStatePending   SchedulingCardState = "Pending"
	SchedulingCardStateCompleted SchedulingCardState = "Completed"
	SchedulingCardStateExpired   SchedulingCardState = "Expired"
)

func (e SchedulingCardState) String() string { return string(e) }
func (e SchedulingCardState) Valid() bool {
	switch e {
	case SchedulingCardStatePending, SchedulingCardStateCompleted, SchedulingCardStateExpired:
		return true
	}
	return false
}

// VenueRevisionReason represents venue rejection reasons.
type VenueRevisionReason string

const (
	VenueRevisionReasonTooFar          VenueRevisionReason = "Too Far"
	VenueRevisionReasonTooExpensive    VenueRevisionReason = "Too Expensive"
	VenueRevisionReasonPreferDifferent VenueRevisionReason = "Prefer Different Place"
)

func (e VenueRevisionReason) String() string { return string(e) }
func (e VenueRevisionReason) Valid() bool {
	switch e {
	case VenueRevisionReasonTooFar, VenueRevisionReasonTooExpensive, VenueRevisionReasonPreferDifferent:
		return true
	}
	return false
}

// MatchLifecycleStatus represents detailed match lifecycle states.
type MatchLifecycleStatus string

const (
	MatchLifecycleStatusConfirmed                   MatchLifecycleStatus = "Confirmed"
	MatchLifecycleStatusScheduling                  MatchLifecycleStatus = "Scheduling"
	MatchLifecycleStatusDateSet                     MatchLifecycleStatus = "Date Set"
	MatchLifecycleStatusDateCompletePendingFeedback MatchLifecycleStatus = "Date Complete Pending Feedback"
	MatchLifecycleStatusDecisionPendingWindow       MatchLifecycleStatus = "Decision Pending Window"
	MatchLifecycleStatusQueued                      MatchLifecycleStatus = "Queued"
	MatchLifecycleStatusClosed                      MatchLifecycleStatus = "Closed"
)

func (e MatchLifecycleStatus) String() string { return string(e) }
func (e MatchLifecycleStatus) Valid() bool {
	switch e {
	case MatchLifecycleStatusConfirmed, MatchLifecycleStatusScheduling, MatchLifecycleStatusDateSet,
		MatchLifecycleStatusDateCompletePendingFeedback, MatchLifecycleStatusDecisionPendingWindow,
		MatchLifecycleStatusQueued, MatchLifecycleStatusClosed:
		return true
	}
	return false
}

// MobilityConstraint represents accessibility needs.
type MobilityConstraint string

const (
	MobilityConstraintWheelchairAccessible MobilityConstraint = "Wheelchair Accessible"
	MobilityConstraintLimitedWalking       MobilityConstraint = "Limited Walking"
	MobilityConstraintNoStairs             MobilityConstraint = "No Stairs"
	MobilityConstraintServiceAnimal        MobilityConstraint = "Service Animal"
	MobilityConstraintNone                 MobilityConstraint = "None"
)

func (e MobilityConstraint) String() string { return string(e) }
func (e MobilityConstraint) Valid() bool {
	switch e {
	case MobilityConstraintWheelchairAccessible, MobilityConstraintLimitedWalking,
		MobilityConstraintNoStairs, MobilityConstraintServiceAnimal, MobilityConstraintNone:
		return true
	}
	return false
}

// DateTypeSubtype represents detailed venue subtypes for TasteBrain.
type DateTypeSubtype string

const (
	// Coffee subtypes
	DateTypeSubtypeCoffee DateTypeSubtype = "coffee"
	DateTypeSubtypeTea    DateTypeSubtype = "tea"
	DateTypeSubtypeBakery DateTypeSubtype = "bakery"

	// Drinks subtypes
	DateTypeSubtypeWine        DateTypeSubtype = "wine"
	DateTypeSubtypeCocktails   DateTypeSubtype = "cocktails"
	DateTypeSubtypeBeerPub     DateTypeSubtype = "beer_pub"
	DateTypeSubtypeWineBar     DateTypeSubtype = "wine_bar"
	DateTypeSubtypeCocktailBar DateTypeSubtype = "cocktail_bar"
	DateTypeSubtypeBar         DateTypeSubtype = "bar"
	DateTypeSubtypeNightClub   DateTypeSubtype = "night_club"

	// Meal subtypes
	DateTypeSubtypeBreakfast         DateTypeSubtype = "breakfast"
	DateTypeSubtypeBrunch            DateTypeSubtype = "brunch"
	DateTypeSubtypeLunch             DateTypeSubtype = "lunch"
	DateTypeSubtypeDinner            DateTypeSubtype = "dinner"
	DateTypeSubtypeDinnerCasual      DateTypeSubtype = "dinner_casual"
	DateTypeSubtypeDinnerSmallPlates DateTypeSubtype = "dinner_small_plates"
	DateTypeSubtypeRestaurant        DateTypeSubtype = "restaurant"
	DateTypeSubtypeTakeaway          DateTypeSubtype = "takeaway"

	// Walk subtypes
	DateTypeSubtypeNeutralWalk       DateTypeSubtype = "neutral_walk"
	DateTypeSubtypePark              DateTypeSubtype = "park"
	DateTypeSubtypeWaterfront        DateTypeSubtype = "waterfront"
	DateTypeSubtypePicnic            DateTypeSubtype = "picnic"
	DateTypeSubtypeNature            DateTypeSubtype = "nature"
	DateTypeSubtypeBeach             DateTypeSubtype = "beach"
	DateTypeSubtypeTouristAttraction DateTypeSubtype = "tourist_attraction"

	// Activity subtypes
	DateTypeSubtypeSauna               DateTypeSubtype = "sauna"
	DateTypeSubtypeSaunaSwim           DateTypeSubtype = "sauna_swim"
	DateTypeSubtypeSwim                DateTypeSubtype = "swim"
	DateTypeSubtypeHiking              DateTypeSubtype = "hiking"
	DateTypeSubtypeIceSkating          DateTypeSubtype = "ice_skating"
	DateTypeSubtypeMiniGolf            DateTypeSubtype = "mini_golf"
	DateTypeSubtypeBowling             DateTypeSubtype = "bowling"
	DateTypeSubtypeArcadeBar           DateTypeSubtype = "arcade_bar"
	DateTypeSubtypeDartsPool           DateTypeSubtype = "darts_pool"
	DateTypeSubtypeAxeThrowing         DateTypeSubtype = "axe_throwing"
	DateTypeSubtypeShootingRange       DateTypeSubtype = "shooting_range"
	DateTypeSubtypeEscapeRoom          DateTypeSubtype = "escape_room"
	DateTypeSubtypeBoardGames          DateTypeSubtype = "board_games"
	DateTypeSubtypeKaraoke             DateTypeSubtype = "karaoke"
	DateTypeSubtypeMuseum              DateTypeSubtype = "museum"
	DateTypeSubtypeGallery             DateTypeSubtype = "gallery"
	DateTypeSubtypeArtExhibit          DateTypeSubtype = "art_exhibit"
	DateTypeSubtypeMovie               DateTypeSubtype = "movie"
	DateTypeSubtypeCinema              DateTypeSubtype = "cinema"
	DateTypeSubtypeConcertLiveMusic    DateTypeSubtype = "concert_live_music"
	DateTypeSubtypeComedyShow          DateTypeSubtype = "comedy_show"
	DateTypeSubtypeSportsGame          DateTypeSubtype = "sports_game"
	DateTypeSubtypeWorkshopClass       DateTypeSubtype = "workshop_class"
	DateTypeSubtypeOtherActivity       DateTypeSubtype = "other_activity"
	DateTypeSubtypeGenericActivity     DateTypeSubtype = "generic_activity"
	DateTypeSubtypeNightClubActivity   DateTypeSubtype = "night_club_activity"
	DateTypeSubtypeArtExhibitOrCulture DateTypeSubtype = "art_exhibit_or_culture"
	DateTypeSubtypeUnknown             DateTypeSubtype = "unknown"
)

func (e DateTypeSubtype) String() string { return string(e) }
func (e DateTypeSubtype) Valid() bool {
	switch e {
	case DateTypeSubtypeCoffee, DateTypeSubtypeTea, DateTypeSubtypeBakery,
		DateTypeSubtypeWine, DateTypeSubtypeCocktails, DateTypeSubtypeBeerPub,
		DateTypeSubtypeWineBar, DateTypeSubtypeCocktailBar, DateTypeSubtypeBar, DateTypeSubtypeNightClub,
		DateTypeSubtypeBreakfast, DateTypeSubtypeBrunch, DateTypeSubtypeLunch, DateTypeSubtypeDinner,
		DateTypeSubtypeDinnerCasual, DateTypeSubtypeDinnerSmallPlates, DateTypeSubtypeRestaurant, DateTypeSubtypeTakeaway,
		DateTypeSubtypeNeutralWalk, DateTypeSubtypePark, DateTypeSubtypeWaterfront, DateTypeSubtypePicnic,
		DateTypeSubtypeNature, DateTypeSubtypeBeach, DateTypeSubtypeTouristAttraction,
		DateTypeSubtypeSauna, DateTypeSubtypeSaunaSwim, DateTypeSubtypeSwim, DateTypeSubtypeHiking,
		DateTypeSubtypeIceSkating, DateTypeSubtypeMiniGolf, DateTypeSubtypeBowling, DateTypeSubtypeArcadeBar,
		DateTypeSubtypeDartsPool, DateTypeSubtypeAxeThrowing, DateTypeSubtypeShootingRange, DateTypeSubtypeEscapeRoom,
		DateTypeSubtypeBoardGames, DateTypeSubtypeKaraoke, DateTypeSubtypeMuseum, DateTypeSubtypeGallery,
		DateTypeSubtypeArtExhibit, DateTypeSubtypeMovie, DateTypeSubtypeCinema, DateTypeSubtypeConcertLiveMusic,
		DateTypeSubtypeComedyShow, DateTypeSubtypeSportsGame, DateTypeSubtypeWorkshopClass, DateTypeSubtypeOtherActivity,
		DateTypeSubtypeGenericActivity, DateTypeSubtypeNightClubActivity, DateTypeSubtypeArtExhibitOrCulture, DateTypeSubtypeUnknown:
		return true
	}
	return false
}

// BookingReminderStatus represents booking reminder lifecycle.
type BookingReminderStatus string

const (
	BookingReminderStatusPending   BookingReminderStatus = "Pending"
	BookingReminderStatusFired     BookingReminderStatus = "Fired"
	BookingReminderStatusDismissed BookingReminderStatus = "Dismissed"
	BookingReminderStatusCompleted BookingReminderStatus = "Completed"
)

func (e BookingReminderStatus) String() string { return string(e) }
func (e BookingReminderStatus) Valid() bool {
	switch e {
	case BookingReminderStatusPending, BookingReminderStatusFired,
		BookingReminderStatusDismissed, BookingReminderStatusCompleted:
		return true
	}
	return false
}

// BookingFailureReason represents booking failure causes.
type BookingFailureReason string

const (
	BookingFailureReasonFullyBooked BookingFailureReason = "Fully Booked"
	BookingFailureReasonNoAnswer    BookingFailureReason = "No Answer"
	BookingFailureReasonClosed      BookingFailureReason = "Closed"
	BookingFailureReasonOther       BookingFailureReason = "Other"
)

func (e BookingFailureReason) String() string { return string(e) }
func (e BookingFailureReason) Valid() bool {
	switch e {
	case BookingFailureReasonFullyBooked, BookingFailureReasonNoAnswer,
		BookingFailureReasonClosed, BookingFailureReasonOther:
		return true
	}
	return false
}

// VenueProposalStatus represents receiver's response to venue proposal.
type VenueProposalStatus string

const (
	VenueProposalStatusPending  VenueProposalStatus = "Pending"
	VenueProposalStatusAccepted VenueProposalStatus = "Accepted"
	VenueProposalStatusRejected VenueProposalStatus = "Rejected"
)

func (e VenueProposalStatus) String() string { return string(e) }
func (e VenueProposalStatus) Valid() bool {
	switch e {
	case VenueProposalStatusPending, VenueProposalStatusAccepted, VenueProposalStatusRejected:
		return true
	}
	return false
}

// AvailabilitySyncMode represents calendar sync types.
type AvailabilitySyncMode string

const (
	AvailabilitySyncModeCalendar AvailabilitySyncMode = "Calendar"
	AvailabilitySyncModeManual   AvailabilitySyncMode = "Manual"
)

func (e AvailabilitySyncMode) String() string { return string(e) }
func (e AvailabilitySyncMode) Valid() bool {
	switch e {
	case AvailabilitySyncModeCalendar, AvailabilitySyncModeManual:
		return true
	}
	return false
}
