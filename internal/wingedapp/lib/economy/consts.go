package economy

const (

	/* category types */

	CategoryTypeAction = "Wings Economy - Action Log"

	/* categories */

	CategoryWingedxWeeklyPayment    = "WingedX - Weekly Payment"
	CategoryWingedxMonthlyPayment   = "WingedX - Monthly Payment"
	CategoryWingedPlusWeeklyPayment = "Winged+ - Weekly Payment"
	CategoryTypeSubscriptionPlan    = "Wings Economy - Subscription Plan"

	/* TODO: deprecate - old economy no longer used */

	CountWingedXPaymentWeek  = 7
	CountWingedXPaymentMonth = 1

	SubscriptionTypeWingedPlus    = "Winged+"
	SubscriptionTypeWingedX       = "WingedX"
	SubscriptionPaymentWeekly     = "Weekly"
	SubscriptionPaymentMonthly    = "Monthly"
	SubscriptionPaymentThreeMonth = "3 Months"
	SubscriptionPaymentSixMonth   = "6 Months"
)

type incrementableCol string

const (
	SentMessages incrementableCol = "Sent Messages"
)

type ActionType string

const (
	/*
		Maps 1:1 with categories.
	*/

	ActionWingedXWeeklyPayment  ActionType = "WingedX - Weekly Payment"
	ActionWingedXMonthlyPayment ActionType = "WingedX - Monthly Payment"

	ActionWingedPlusWeeklyPayment     ActionType = "Winged+ - Weekly Payment"
	ActionWingedPlusMonthlyPayment    ActionType = "Winged+ - Monthly Payment"
	ActionWingedPlusThreeMonthPayment ActionType = "Winged+ - 3 Month Payment"
	ActionWingedPlusSixMonthPayment   ActionType = "Winged+ - 6 Month Payment"

	/* Daily Check-in */

	ActionDailyCheckIn ActionType = "Daily Check-In"

	/* Streak Milestones - Per spec: wings granted ONLY at milestones */

	ActionStreak7Day  ActionType = "Streak - 7 Day Milestone"
	ActionStreak30Day ActionType = "Streak - 30 Day Milestone"

	/* Referral */

	// When invited user performs first paid action (connect or schedule)
	// Referrer gets 4 wings, invitee gets nothing per MVP spec
	ActionReferralComplete ActionType = "Referral - Friend Complete"

	/* Date Completion */

	// When user confirms they attended a scheduled date (did_meet: true)
	ActionAttendDate ActionType = "Attend a Date"

	/* Message Spending */

	// When user sends messages (every 5 messages costs 1 wing)
	ActionSendMessage ActionType = "Send Message"
)

// Streak milestone rewards - Per spec: check-in does NOT grant wings directly
// Wings are granted ONLY through streak milestones
const (
	Streak7DayWings  = 2 // +2 wings at 7 consecutive days
	Streak30DayWings = 6 // +6 wings at 30 consecutive days
)

// ReferralBonusWings is the amount of wings earned by referrer per successful referral
// Per MVP spec: +4 wings to referrer only (invitee gets nothing)
const ReferralBonusWings = 4

// AttendDateWings is the amount of wings earned for attending a scheduled date
const AttendDateWings = 1

// SendMessageThreshold is the number of messages before 1 wing is deducted
const SendMessageThreshold = 5

// SendMessageWingsCost is the amount of wings deducted per threshold
const SendMessageWingsCost = 1

// EarnedWingsExpiryDays is the number of days before earned wings expire
const EarnedWingsExpiryDays = 30
